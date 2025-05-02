package models

import (
	"time"
)

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
	Active                   bool
	ValidatedBy              int
	Validator                Admin `gorm:"foreignKey:ValidatedBy;references:AdminID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type UserOperation struct {
	ID              int `gorm:"primaryKey"`
	UserID          int
	UserRef         User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Operation       GeneralCrudOperation
	OperationStatus GeneralTransactionState
	PerformedAt     time.Time `gorm:"autoCreateTime"`
}
