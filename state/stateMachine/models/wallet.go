package models

import (
	"raft/utils"
	"time"
)

type TransferFD struct {
	Wallet2    *int
	Wallet2Ref *Wallet `gorm:"foreignKey:Wallet2;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type Wallet struct {
	WalletID  int `gorm:"primaryKey"`
	UserID    int
	UserRef   User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Balance   int64     `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type WalletOperation struct {
	ID             int `gorm:"primaryKey"`
	Type           utils.WalletAction
	Amount         int64
	Timestamp      time.Time               `gorm:"autoCreateTime"`
	Status         utils.TransactionStatus `gorm:"default:'pending'"`
	Wallet1        int
	Wallet1Ref     Wallet `gorm:"foreignKey:Wallet1;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TranferFundsFD *TransferFD
}
