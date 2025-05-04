package models

type CreateAdminFormData struct {
	FirstName, LastName, Email, Password string
}

type ValidateAccountFornData struct {
	UserID int
}

type Admin struct {
	AdminID        int `gorm:"primaryKey"`
	FirstName      string
	LastName       string
	HashedPassword string
	Email          string `gorm:"unique"`
}

type AdminOperation struct {
	ID                      int `gorm:"primaryKey"`
	AdminID                 int
	AdminRef                Admin `gorm:"foreignKey:AdminID;references:AdminID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreateAdminFormData     *CreateAdminFormData
	ValidateAccountFormData *ValidateAccountFornData
	OperationStatus         GeneralTransactionState
	OperationType           AdminOperations
}
