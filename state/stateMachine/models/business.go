package models

import (
	"raft/utils"
	"time"
)

type Wallet struct {
	WalletID  int `gorm:"primaryKey"`
	UserID    int
	UserRef   User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Balance   int64     `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Admin struct {
	AdminID                             int `gorm:"primaryKey"`
	FirstName, LastName, HashedPassword string
	Email                               string `gorm:"unique"`
	Active                              bool   `gorm:"default:true"`
}

type WalletOperation struct {
	ID         int `gorm:"primaryKey"`
	Type       utils.WalletAction
	Amount     int64
	Timestamp  time.Time
	Status     utils.TransactionStatus
	Wallet1    int
	Wallet1Ref Wallet `gorm:"foreignKey:Wallet1;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Wallet2    *int
	Wallet2Ref Wallet `gorm:"foreignKey:Wallet2;references:WalletID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type User struct {
	UserID                   int `gorm:"primaryKey"`
	FirstName                string
	LastName                 string
	HashedPassword           string
	Email                    string  `gorm:"unique"`
	Rating                   float32 `gorm:"default:0.0"`
	DateOfBirth              time.Time
	IdentificationNumber     string `gorm:"unique"`
	IdentificationImageFront string
	IdentificationImageBack  string
	ValidatedBy              *int
	ValidatorRef             Admin     `gorm:"foreignKey:ValidatedBy;references:AdminID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt                time.Time `gorm:"autoCreateTime"`
	UpdatedAt                time.Time
	Active                   bool `gorm:"default:true"`
}
