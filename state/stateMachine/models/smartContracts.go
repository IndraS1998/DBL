package models

import (
	"raft/utils"
	"time"
)

type Escrow struct {
	ID           int `gorm:"primaryKey"`
	WalletID     int
	WalletRef    Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TargetAmount int64
	CreatedAt    time.Time               `gorm:"autoCreateTime"`
	Status       utils.TransactionStatus `gorm:"default:'pending'"`
}

type Loan struct {
	ID              int `gorm:"primaryKey"`
	WalletID        int
	WalletRef       Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Amount          int64
	Rate            float32
	DueDate         time.Time
	RepaymentAmount int64
	Status          utils.TransactionStatus `gorm:"default:'pending'"`
}

type LoanContribution struct {
	ID        int `gorm:"primaryKey"`
	LoanID    int
	LoanRef   Loan `gorm:"foreignKey:LoanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	WalletID  int
	WalletRef Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Amount    int64
	Time      time.Time `gorm:"autoCreateTime"`
}
