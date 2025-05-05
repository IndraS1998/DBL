package models

import (
	"raft/utils"
	"time"
)

type EscrowFormData struct {
	TargetAmount int64
	WalletID     int
}

type Escrow struct {
	ID           int `gorm:"primaryKey"`
	WalletID     int
	WalletRef    Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TargetAmount int64
	CreatedAt    time.Time               `gorm:"autoCreateTime"`
	Status       utils.TransactionStatus `gorm:"default:'pending'"`
}
