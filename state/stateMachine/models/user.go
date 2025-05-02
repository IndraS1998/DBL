package models

import (
	"time"
)

type CreateAccountFormData struct {
	FirstName, LastName, Email, Password string
	DateOfBirth                          time.Time
	IdentificationNumber                 string
	IdentificationImageFront             string
	IdentificationImageBack              string
}

type UpdatePasswordFormData struct {
	PreviousPassword, NewPassWord string
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
	ValidatorRef             Admin `gorm:"foreignKey:ValidatedBy;references:AdminID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type UserOperation struct {
	ID                     int `gorm:"primaryKey"`
	UserID                 int
	UserRef                User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreateAccountFormData  *CreateAccountFormData
	UpdatePasswordFormData *UpdatePasswordFormData
	Operation              UserOperations
	OperationStatus        GeneralTransactionState
	PerformedAt            time.Time `gorm:"autoCreateTime"`
}
