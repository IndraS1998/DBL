package models

type Admin struct {
	AdminID        int `gorm:"primaryKey"`
	FirstName      string
	LastName       string
	HashedPassword string
	Email          string `gorm:"unique"`
}
