package models

import (
	"time"
)

type EscrowFormData struct {
	TargetAmount int64
	WalletID     int
}

type Escrow struct {
	ID                   int `gorm:"primaryKey"`
	WalletID             int
	WalletRef            Wallet `gorm:"foreignKey:WalletID;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TargetAmount         int64
	CreationOperation    int
	CreationOperationRef EscrowOperation         `gorm:"foreignKey:CreationOperation;references:ID;constraint:OnUpdate:CASCADE,onDelete:SET NULL"`
	CreatedAt            time.Time               `gorm:"autoCreateTime"`
	Status               GeneralTransactionState `gorm:"default:'pending'"`
}

type EscrowOperation struct {
	ID              int `gorm:"primaryKey"`
	Operation       GeneralCrudOperation
	OperationStatus GeneralTransactionState `gorm:"default:'pending'"`
	EscrowFD        *EscrowFormData
	PerformedAt     time.Time `gorm:"autoCreateTime"`
}

// TODO instad of the operation holding a ref o the escrow, the reverse should happen
