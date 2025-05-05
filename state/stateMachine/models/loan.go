package models

import (
	"raft/utils"
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
	Status          utils.TransactionStatus `gorm:"default:'pending'"`
}
