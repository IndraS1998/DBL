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

// Deposit adds the specified amount to the wallet's balance.
// Returns an error only if the wallet is not found or business logic fails.
func (sm *StateMachine) Deposit(walletID int, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}

	var wallet models.Wallet
	// Find the wallet
	if err := sm.DB.First(&wallet, walletID).Error; err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	// Perform atomic update in a transaction
	return sm.DB.Transaction(func(tx *gorm.DB) error {
		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// Optional: log operation (if you have WalletOperation or Transaction table)
		op := models.WalletOperation{
			Wallet1:   walletID,
			Amount:    amount,
			Type:      "deposit",
			Timestamp: time.Now(),
			Status:    "success",
		}
		if err := tx.Create(&op).Error; err != nil {
			return fmt.Errorf("failed to log wallet operation: %w", err)
		}

		return nil
	})
}
