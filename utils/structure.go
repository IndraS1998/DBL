package utils

import "time"

type Payload interface {
	GetRefTable() RefTable
}

// payloads
// payloads for append entries
type UserPayload struct {
	FirstName, LastName, HashedPassword, Email                                             string
	DateOfBirth                                                                            time.Time
	IdentificationNumber, IdentificationImageFront, IdentificationImageBack, PrevPW, NewPW string
	UserID                                                                                 int
	Action                                                                                 UserAction // create, update, delete
}

func (up UserPayload) GetRefTable() RefTable {
	return RefUser
}

type AdminPayload struct {
	FirstName, LastName, HashedPassword, Email string
	AdminID, UserId                            int
	Action                                     AdminAction
}

func (ap AdminPayload) GetRefTable() RefTable {
	return RefAdmin
}

type WalletOperationPayload struct {
	Wallet1, Wallet2 int
	Amount           int64
	Action           WalletAction
}

func (wp WalletOperationPayload) GetRefTable() RefTable {
	return RefWallet
}
