package models

import (
	"time"
)

type CreateUserAccountPayload struct {
	FirstName                string
	LastName                 string
	HashedPassword           string
	Email                    string
	DateOfBirth              time.Time
	IdentificationNumber     string `gorm:"unique"`
	IdentificationImageFront string
	IdentificationImageBack  string
}
type UpdateUserPasswordPayload struct {
	PrevPassword, NewPassword string
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
