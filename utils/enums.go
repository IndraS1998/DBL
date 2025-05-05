package utils

// General status for any transaction
type TransactionStatus string

const (
	TxPending TransactionStatus = "pending"
	TxSuccess TransactionStatus = "success"
	TxFailed  TransactionStatus = "failed"
)

// Tables or operation domains
type RefTable string

const (
	RefWallet RefTable = "wallet"
	RefUser   RefTable = "user"
	RefAdmin  RefTable = "admin"
)

// CRUD operations
type CrudOperation string

const (
	CrudCreate CrudOperation = "create"
	CrudUpdate CrudOperation = "update"
	CrudDelete CrudOperation = "delete"
)

// Wallet-specific actions
type WalletAction string

const (
	WalletDeposit  WalletAction = "deposit"
	WalletWithdraw WalletAction = "withdraw"
	WalletTransfer WalletAction = "transfer"
)

// User-specific actions
type UserAction string

const (
	UserCreateAccount  UserAction = "create_account"
	UserUpdatePassword UserAction = "update_password"
	UserCreateWallet   UserAction = "create_wallet"
	UserDeleteAccount  UserAction = "delete_account"
)

// Admin-specific actions
type AdminAction string

const (
	AdminCreateAccount AdminAction = "create_admin_account"
	AdminValidateUser  AdminAction = "validate_user"
)
