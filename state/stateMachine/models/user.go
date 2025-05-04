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

type UpdateFormData struct {
	PreviousPassword, NewPassWord string
	UserID                        int
	UserRef                       User `gorm:"foreignKey:UserID;references:UserID;constriant:OnUpdate:CASCASE,onDelete:SET NULL"`
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
	CreationOperation        int
	CreationOperationRef     *UserOperation `gorm:"foreignKey:CreationOperation;references:ID;constrint:OnUpdate:CASCASE,OnDelete:SET NULL"`
}

type UserOperation struct {
	ID                    int `gorm:"primaryKey"`
	CreateAccountFormData *CreateAccountFormData
	UpdateFormData        *UpdateFormData
	Operation             UserOperations
	OperationStatus       GeneralTransactionState
	PerformedAt           time.Time `gorm:"autoCreateTime"`
}
