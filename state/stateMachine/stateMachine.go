package stateMachine

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"raft/state/stateMachine/models"
	"raft/utils"
)

var (
	defaultSM *StateMachine
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
	defaultSM = &StateMachine{DB: db}
	fmt.Println("successfully initialized state machine")
	return defaultSM, nil
}

// ordinary get operations

// GetUserByID return the user information
func GetUserByID(userID int) (*models.User, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var user models.User
	if err := defaultSM.DB.First(&user, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("unable to ge the user: %w", err)
	}
	return &user, nil
}

func CountValidatedUsers() (int64, error) {
	var count int64
	err := defaultSM.DB.Model(&models.User{}).Where("active = ?", true).Count(&count).Error
	return count, err
}

func CountWalletOperationsBetween(start, end time.Time) (int64, error) {
	var count int64
	err := defaultSM.DB.Model(&models.WalletOperation{}).
		Where("timestamp >= ? AND timestamp < ?", start, end).
		Count(&count).Error
	return count, err
}

func SumWalletOperationAmountsBetween(start, end time.Time) (float64, error) {
	var total float64
	err := defaultSM.DB.Model(&models.WalletOperation{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("timestamp >= ? AND timestamp < ?", start, end).
		Scan(&total).Error
	return total, err
}

func CountWallets() (int64, error) {
	var count int64
	err := defaultSM.DB.Model(&models.Wallet{}).Count(&count).Error
	return count, err
}

func GetMostRecentWalletOperations(limit int) ([]*models.WalletOperation, error) {
	var operations []*models.WalletOperation
	err := defaultSM.DB.Order("timestamp DESC").
		Limit(limit).
		Find(&operations).Error
	return operations, err
}

func GetUserByEmail(email string) (*models.User, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var user models.User
	if err := defaultSM.DB.First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetAdminInfo returns an admin instace from the db
func GetAdminInfo(adminID int) (*models.Admin, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var admin models.Admin
	if err := defaultSM.DB.First(&admin, "admin_id = ?", adminID).Error; err != nil {
		return nil, fmt.Errorf("unable to get admin: %w", err)
	}
	return &admin, nil
}

func GetAdminByEmail(email string) (*models.Admin, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var admin models.Admin
	if err := defaultSM.DB.First(&admin, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	return &admin, nil
}

// GetWallet info ...
func GetWallet(walletID int) (*models.Wallet, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var wallet models.Wallet
	if err := defaultSM.DB.First(&wallet, "wallet_id = ?", walletID).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func GetWallets(userID int) ([]*models.Wallet, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var wallets []*models.Wallet
	if err := defaultSM.DB.Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("unable to get wallets: %w", err)
	}
	return wallets, nil
}

func GetUsers() ([]*models.User, error) {
	if defaultSM == nil {
		return nil, fmt.Errorf("state machine not yet initialized")
	}
	var users []*models.User
	if err := defaultSM.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
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
					"active":       true,
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
