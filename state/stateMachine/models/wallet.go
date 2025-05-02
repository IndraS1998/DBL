package models

import (
	"time"
)

type Wallet struct {
	WalletID  int `gorm:"primaryKey"`
	UserID    int
	UserRef   User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Balance   int64
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type WalletOperation struct {
	ID         int `gorm:"primaryKey"`
	Type       WalletOperationType
	Amount     int64
	Timestamp  time.Time               `gorm:"autoCreateTime"`
	Status     GeneralTransactionState `gorm:"default:'pending'"`
	Wallet1    int
	Wallet1Ref Wallet `gorm:"foreignKey:Wallet1;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Wallet2    *int
	Wallet2Ref *Wallet `gorm:"foreignKey:Wallet2;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

//getbalance, transfer, deposit, withdraw
