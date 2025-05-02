package stateMachine

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"raft/state/stateMachine/models"
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
func (sm *StateMachine) ApplyWalletOperation(opID string) (*models.WalletOperation, error) {
	var op models.WalletOperation

	// Load the operation by ID
	if err := sm.DB.First(&op, "operation_id = ?", opID).Error; err != nil {
		return nil, fmt.Errorf("failed to find wallet operation: %w", err)
	}

	op.Timestamp = time.Now()

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		var w1 models.Wallet
		if err := tx.First(&w1, "wallet_id = ?", op.Wallet1).Error; err != nil {
			op.Status = models.OperationFailed
			return fmt.Errorf("wallet1 not found: %w", err)
		}

		switch op.Type {
		case models.Deposit:
			w1.Balance += op.Amount

		case models.Withdraw:
			if w1.Balance < op.Amount {
				op.Status = models.OperationFailed
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= op.Amount

		case models.Transfer:
			if op.Wallet2 == nil {
				op.Status = models.OperationFailed
				return fmt.Errorf("wallet2 is required for transfer")
			}
			var w2 models.Wallet
			if err := tx.First(&w2, "wallet_id = ?", *op.Wallet2).Error; err != nil {
				op.Status = models.OperationFailed
				return fmt.Errorf("wallet2 not found: %w", err)
			}
			if w1.Balance < op.Amount {
				op.Status = models.OperationFailed
				return fmt.Errorf("insufficient funds")
			}
			w1.Balance -= op.Amount
			w2.Balance += op.Amount

			if err := tx.Save(&w2).Error; err != nil {
				op.Status = models.OperationFailed
				return fmt.Errorf("failed to update wallet2: %w", err)
			}

		default:
			op.Status = models.OperationFailed
			return fmt.Errorf("unsupported operation type: %v", op.Type)
		}

		if err := tx.Save(&w1).Error; err != nil {
			op.Status = models.OperationFailed
			return fmt.Errorf("failed to update wallet1: %w", err)
		}

		op.Status = models.OperationSuccess
		if err := tx.Save(&op).Error; err != nil {
			return fmt.Errorf("failed to update operation status: %w", err)
		}

		return nil
	})

	return &op, err
}

// ApplyUserOperation performs user operations withdraw, deposit,transfer
// ApplyUserOperation applies a persisted user operation by its ID and updates its status accordingly.
func (sm *StateMachine) ApplyUserOperation(opID string) (*models.UserOperation, error) {
	var op models.UserOperation

	// Load operation
	if err := sm.DB.Preload("UserRef").First(&op, "id = ?", opID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user operation: %w", err)
	}

	op.PerformedAt = time.Now()

	var opErr error

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		switch op.Operation {

		case models.CreateAccount:
			if op.CreateAccountFormData == nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("user data is required")
				return opErr
			}
			user := models.User{
				FirstName:                op.CreateAccountFormData.FirstName,
				LastName:                 op.CreateAccountFormData.LastName,
				Email:                    op.CreateAccountFormData.Email,
				HashedPassword:           op.CreateAccountFormData.Password, // hash before use!
				DateOfBirth:              op.CreateAccountFormData.DateOfBirth,
				IdentificationNumber:     op.CreateAccountFormData.IdentificationNumber,
				IdentificationImageFront: op.CreateAccountFormData.IdentificationImageFront,
				IdentificationImageBack:  op.CreateAccountFormData.IdentificationImageBack,
			}
			if err := tx.Create(&user).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("failed to create user: %w", err)
				return opErr
			}
			op.OperationStatus = models.OperationSuccess

		case models.CreateWallet:
			wallet := models.Wallet{UserID: op.UserID}
			if err := tx.Create(&wallet).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("failed to create wallet: %w", err)
				return opErr
			}
			op.OperationStatus = models.OperationSuccess

		case models.UpdatePassword:
			if op.UpdatePasswordFormData == nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("password data is required")
				return opErr
			}
			if op.UserRef.HashedPassword != op.UpdatePasswordFormData.PreviousPassword {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("invalid password supplied")
				return opErr
			}
			if err := tx.Model(&models.User{}).
				Where("user_id = ?", op.UserID).
				Update("hashed_password", op.UpdatePasswordFormData.NewPassWord).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("unable to update password: %w", err)
				return opErr
			}
			op.OperationStatus = models.OperationSuccess

		case models.DeleteAccount:
			if err := tx.Where("user_id = ?", op.UserID).Delete(&models.User{}).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				opErr = fmt.Errorf("unable to delete user: %w", err)
				return opErr
			}
			op.OperationStatus = models.OperationSuccess

		default:
			op.OperationStatus = models.OperationFailed
			opErr = fmt.Errorf("invalid operation type: %v", op.Operation)
			return opErr
		}

		// Update operation status and performed timestamp
		if err := tx.Model(&models.UserOperation{}).
			Where("id = ?", op.ID).
			Updates(map[string]interface{}{
				"operation_status": op.OperationStatus,
				"performed_at":     op.PerformedAt,
			}).Error; err != nil {
			return fmt.Errorf("failed to update operation status: %w", err)
		}

		return nil
	})

	return &op, err
}

// ApplyAdminOperations commits a previously persisted admin operation and updates its status accordingly.
func (sm *StateMachine) ApplyAdminOperations(opID string) (*models.AdminOperation, error) {
	var op models.AdminOperation

	// Load the operation from DB
	if err := sm.DB.First(&op, "operation_id = ?", opID).Error; err != nil {
		return nil, fmt.Errorf("failed to find admin operation: %w", err)
	}

	err := sm.DB.Transaction(func(tx *gorm.DB) error {
		switch op.OperationType {
		case models.CreateAdminAccount:
			if op.CreateAdminFormData == nil {
				op.OperationStatus = models.OperationFailed
				return fmt.Errorf("admin data was required")
			}
			admin := models.Admin{
				FirstName:      op.CreateAdminFormData.FirstName,
				LastName:       op.CreateAdminFormData.LastName,
				HashedPassword: op.CreateAdminFormData.Password,
				Email:          op.CreateAdminFormData.Email,
			}
			if err := tx.Create(&admin).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				return fmt.Errorf("failed to create admin: %w", err)
			}
			op.OperationStatus = models.OperationSuccess

		case models.ValidateUserAccount:
			if op.ValidateAccountFormData == nil {
				op.OperationStatus = models.OperationFailed
				return fmt.Errorf("form data was required")
			}
			if err := tx.Model(&models.User{}).
				Where("user_id = ?", op.ValidateAccountFormData.UserID).
				Update("validated_by", op.AdminID).Error; err != nil {
				op.OperationStatus = models.OperationFailed
				return fmt.Errorf("failed to validate user: %w", err)
			}
			op.OperationStatus = models.OperationSuccess

		default:
			op.OperationStatus = models.OperationFailed
			return fmt.Errorf("invalid admin operation type: %v", op.OperationType)
		}

		// Save the operation's final status
		if err := tx.Save(&op).Error; err != nil {
			return fmt.Errorf("failed to update admin operation status: %w", err)
		}

		return nil
	})

	return &op, err
}
