package models

import (
	"time"
)

type LoanContribution struct {
	ID        int `gorm:"primaryKey"`
	LoanID    int
	LoanRef   Loan `gorm:"foreignKey:LoanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	WalletID  int
	WalletRef Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Amount    int64
	Time      time.Time `gorm:"autoCreateTime"`
}

type LoanContributionOperation struct {
	ID                  int `gorm:"primaryKey"`
	LoanContributionID  int
	LoanContributionRef LoanContribution `gorm:"foreignKey:LoanContributionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	OperationType       GeneralCrudOperation
	OperationStatus     GeneralTransactionState `gorm:"default:'pending'"`
	OperationTime       time.Time               `gorm:"autoCreateTime"`
}
