package models

type CreateAdminAccountPayload struct {
	FirstName, LastName, HashedPassword, Email string
}
type ValidateUserAccountPayload struct {
	UserID int
}
type Admin struct {
	AdminID                             int `gorm:"primaryKey"`
	FirstName, LastName, HashedPassword string
	Email                               string `gorm:"unique"`
	Active                              bool   `gorm:"default:true"`
}
