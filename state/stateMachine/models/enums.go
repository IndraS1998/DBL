package models

type GeneralTransactionState string

const (
	Pending GeneralTransactionState = "pending"
	Success GeneralTransactionState = "success"
	Failed  GeneralTransactionState = "failed"
)

type WalletOperationType string

const (
	Deposit  WalletOperationType = "deposit"
	Withdraw WalletOperationType = "withdraw"
	Transfer WalletOperationType = "transfer"
)

type GeneralCrudOperation string

const (
	Create GeneralCrudOperation = "create"
	Update GeneralCrudOperation = "update"
	Delete GeneralCrudOperation = "delete"
)
