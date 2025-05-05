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
