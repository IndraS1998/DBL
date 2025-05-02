package models

import (
	"time"
)

type Escrow struct {
	ID           int `gorm:"primaryKey"`
	WalletID     int
	WalletRef    Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TargetAmount int64
	CreatedAt    time.Time               `gorm:"autoCreateTime"`
	Status       GeneralTransactionState `gorm:"default:'pending'"`
}

type EscrowOperation struct {
	ID              int `gorm:"primaryKey"`
	EscrowID        int
	EscrowRef       Escrow `gorm:"foreignKey:EscrowID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Operation       GeneralCrudOperation
	OperationStatus GeneralTransactionState
	PerformedAt     time.Time `gorm:"autoCreateTime"`
}
