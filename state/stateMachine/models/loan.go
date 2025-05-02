package models

import (
	"time"
)

type Loan struct {
	ID              int `gorm:"primaryKey"`
	WalletID        int
	WalletRef       Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Amount          int64
	Rate            float32
	DueDate         time.Time
	RepaymentAmount int64
	Status          GeneralTransactionState `gorm:"default:'pending'"`
}

type LoanOperation struct {
	ID              int `gorm:"primaryKey"`
	LoanID          int
	LoanRef         Loan `gorm:"foreignKey:LoanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	OperationType   GeneralCrudOperation
	OperationStatus GeneralTransactionState
	OperationDate   time.Time `gorm:"autoCreateTime"`
}
