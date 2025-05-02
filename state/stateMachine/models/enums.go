package models

type GeneralTransactionState string

const (
	OperationPending GeneralTransactionState = "pending"
	OperationSuccess GeneralTransactionState = "success"
	OperationFailed  GeneralTransactionState = "failed"
)

type WalletOperationType string

const (
	Deposit  WalletOperationType = "deposit"
	Withdraw WalletOperationType = "withdraw"
	Transfer WalletOperationType = "transfer"
)

type GeneralCrudOperation string

const (
	CRUDCreate GeneralCrudOperation = "create"
	CRUDUpdate GeneralCrudOperation = "update"
	CRUDDelete GeneralCrudOperation = "delete"
)

type UserOperations string

const (
	CreateAccount  UserOperations = "create"
	UpdatePassword UserOperations = "password"
	CreateWallet   UserOperations = "wallet_create"
	DeleteAccount  UserOperations = "delete"
)

type AdminOperations string

const (
	CreateAdminAccount  AdminOperations = "create"
	ValidateUserAccount AdminOperations = "validate"
)
