package stateMachine

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"raft/state/stateMachine/models"
	"raft/utils"
)

type StateMachine struct {
	DB *gorm.DB
}

// initialize the database and auto migrate
func InitStateMachine(path string) (*StateMachine, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite DB at %s: %w", path, err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Admin{}, &models.User{}, &models.Wallet{}, &models.WalletOperation{})
	if err != nil {
		return nil, fmt.Errorf("failed automigrate %w", err)
	}

	return &StateMachine{DB: db}, nil
}

// ApplyWalletOperation performs balace mutation on a wallet
// ApplyWalletOperation applies a persisted wallet operation and updates its status.
func (sm *StateMachine) ApplyWalletOperation(wallet1 int, wallet2 int, amount int64, a utils.WalletAction) error {

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		// get first wallet
		var w1 models.Wallet
		if err := tx.First(&w1, "wallet_id = ?", wallet1).Error; err != nil {
			return fmt.Errorf("wallet1 not found: %w", err)
		}
		//perform wallet actions
		switch a {
		case utils.WalletDeposit:
			w1.Balance += amount

		case utils.WalletWithdraw:
			if w1.Balance < amount {
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= amount

		case utils.WalletTransfer:
			if wallet2 < 0 {
				return fmt.Errorf("wallet2 is required for transfer")
			}
			var w2 models.Wallet
			if err := tx.First(&w2, "wallet_id = ?", wallet2).Error; err != nil {
				return fmt.Errorf("wallet2 not found: %w", err)
			}
			if w1.Balance < amount {
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= amount
			w2.Balance += amount

			if err := tx.Save(&w2).Error; err != nil {
				return fmt.Errorf("failed to update wallet2: %w", err)
			}

		default:
			return fmt.Errorf("unsupported operation type: %s", a)
		}

		if err := tx.Save(&w1).Error; err != nil {
			return fmt.Errorf("failed to update wallet1: %w", err)
		}
		//TODO equally save transaction
		return nil
	})

	return err
}

// ApplyUserOperation performs user, UserID is set to -1 if not required such as create
func (sm *StateMachine) ApplyUserOperation(UserID int, a utils.UserAction, cu *models.CreateUserAccountPayload, uu *models.UpdateUserPasswordPayload) error {
	err := sm.DB.Transaction(func(tx *gorm.DB) error {

		switch a {

		case utils.UserCreateAccount:
			if cu == nil {
				return fmt.Errorf("user data is required")
			}
			user := models.User{
				FirstName:                cu.FirstName,
				LastName:                 cu.LastName,
				Email:                    cu.Email,
				HashedPassword:           cu.HashedPassword, // hash before use!
				DateOfBirth:              cu.DateOfBirth,
				IdentificationNumber:     cu.IdentificationNumber,
				IdentificationImageFront: cu.IdentificationImageFront,
				IdentificationImageBack:  cu.IdentificationImageBack,
			}
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case utils.UserCreateWallet:
			wallet := models.Wallet{UserID: UserID}
			if err := tx.Create(&wallet).Error; err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
		case utils.UserUpdatePassword:
			if uu == nil {
				return fmt.Errorf("password data is required")
			}
			var user models.User
			if err := tx.First(&user, "user_id = ?", UserID).Error; err != nil {
				return fmt.Errorf("unable to get user:%w", err)
			}
			if user.HashedPassword != uu.PrevPassword {
				return fmt.Errorf("invalid password supplied")
			}
			user.HashedPassword = uu.NewPassword
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("unable to update password: %w", err)
			}
		case utils.UserDeleteAccount:
			if err := tx.Where("user_id = ?", UserID).Delete(&models.User{}).Error; err != nil {
				return fmt.Errorf("unable to delete user: %w", err)
			}

		default:
			return fmt.Errorf("invalid operation type: %s", a)
		}
		return nil
	})

	return err
}

// ApplyAdminOperations commits logs related to admin operation. the id is -1 if it is not required
func (sm *StateMachine) ApplyAdminOperations(AdminID int, ca *models.CreateAdminAccountPayload, vu *models.ValidateUserAccountPayload, a utils.AdminAction) error {

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		switch a {
		case utils.AdminCreateAccount:
			if ca == nil {
				return fmt.Errorf("admin data was required")
			}
			admin := models.Admin{
				FirstName:      ca.FirstName,
				LastName:       ca.LastName,
				HashedPassword: ca.HashedPassword,
				Email:          ca.Email,
			}
			if err := tx.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to create admin: %w", err)
			}
		case utils.AdminValidateUser:
			if vu == nil {
				return fmt.Errorf("form data was required")
			}
			if err := tx.Model(&models.User{}).
				Where("user_id = ?", vu.UserID).
				Updates(map[string]interface{}{
					"validated_by": AdminID,
					"updated_at":   time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to validate user: %w", err)
			}

		default:
			return fmt.Errorf("invalid admin operation type: %s", a)
		}

		return nil
	})

	return err
}
