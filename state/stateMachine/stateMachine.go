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
func (sm *StateMachine) ApplyWalletOperation(walletPayload utils.WalletOperationPayload) error {

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		// get first wallet
		var w1 models.Wallet
		if err := tx.First(&w1, "wallet_id = ?", walletPayload.Wallet1).Error; err != nil {
			return fmt.Errorf("wallet1 not found: %w", err)
		}
		//perform wallet actions
		switch walletPayload.Action {
		case utils.WalletDeposit:
			w1.Balance += walletPayload.Amount

		case utils.WalletWithdraw:
			if w1.Balance < walletPayload.Amount {
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= walletPayload.Amount

		case utils.WalletTransfer:
			if walletPayload.Wallet2 < 0 {
				return fmt.Errorf("wallet2 is required for transfer")
			}
			var w2 models.Wallet
			if err := tx.First(&w2, "wallet_id = ?", walletPayload.Wallet2).Error; err != nil {
				return fmt.Errorf("wallet2 not found: %w", err)
			}
			if w1.Balance < walletPayload.Amount {
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= walletPayload.Amount
			w2.Balance += walletPayload.Amount

			if err := tx.Save(&w2).Error; err != nil {
				return fmt.Errorf("failed to update wallet2: %w", err)
			}

		default:
			return fmt.Errorf("unsupported operation type: %s", walletPayload.Action)
		}

		if err := tx.Save(&w1).Error; err != nil {
			return fmt.Errorf("failed to update wallet1: %w", err)
		}
		walletOperation := models.WalletOperation{
			Wallet1:   walletPayload.Wallet1,
			Wallet2:   &walletPayload.Wallet2,
			Amount:    walletPayload.Amount,
			Type:      walletPayload.Action,
			Timestamp: time.Now(),
			Status:    utils.TxSuccess,
		}
		if errop := tx.Create(&walletOperation).Error; errop != nil {
			return fmt.Errorf("failed to create wallet operation: %w", errop)
		}
		return nil
	})

	return err
}

// ApplyUserOperation performs user, UserID is set to -1 if not required such as create
func (sm *StateMachine) ApplyUserOperation(userPayload utils.UserPayload) error {
	err := sm.DB.Transaction(func(tx *gorm.DB) error {

		switch userPayload.Action {

		case utils.UserCreateAccount:
			user := models.User{
				FirstName:                userPayload.FirstName,
				LastName:                 userPayload.LastName,
				Email:                    userPayload.Email,
				HashedPassword:           userPayload.HashedPassword, // hash before use!
				DateOfBirth:              userPayload.DateOfBirth,
				IdentificationNumber:     userPayload.IdentificationNumber,
				IdentificationImageFront: userPayload.IdentificationImageFront,
				IdentificationImageBack:  userPayload.IdentificationImageBack,
			}
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case utils.UserCreateWallet:
			wallet := models.Wallet{UserID: userPayload.UserID}
			if err := tx.Create(&wallet).Error; err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
		case utils.UserUpdatePassword:
			var user models.User
			if err := tx.First(&user, "user_id = ?", userPayload.UserID).Error; err != nil {
				return fmt.Errorf("unable to get user:%w", err)
			}
			if user.HashedPassword != userPayload.PrevPW {
				return fmt.Errorf("invalid password supplied")
			}
			user.HashedPassword = userPayload.NewPW
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("unable to update password: %w", err)
			}
		case utils.UserDeleteAccount:
			if err := tx.Where("user_id = ?", userPayload.UserID).Delete(&models.User{}).Error; err != nil {
				return fmt.Errorf("unable to delete user: %w", err)
			}

		default:
			return fmt.Errorf("invalid operation type: %s", userPayload.Action)
		}
		return nil
	})

	return err
}

// ApplyAdminOperations commits logs related to admin operation. the id is -1 if it is not required
func (sm *StateMachine) ApplyAdminOperations(adminPayload utils.AdminPayload) error {

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		switch adminPayload.Action {
		case utils.AdminCreateAccount:
			admin := models.Admin{
				FirstName:      adminPayload.FirstName,
				LastName:       adminPayload.LastName,
				HashedPassword: adminPayload.HashedPassword,
				Email:          adminPayload.Email,
			}
			if err := tx.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to create admin: %w", err)
			}
		case utils.AdminValidateUser:
			if err := tx.Model(&models.User{}).
				Where("user_id = ?", adminPayload.UserId).
				Updates(map[string]interface{}{
					"validated_by": adminPayload.AdminID,
					"updated_at":   time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to validate user: %w", err)
			}

		default:
			return fmt.Errorf("invalid admin operation type: %s", adminPayload.Action)
		}

		return nil
	})

	return err
}
