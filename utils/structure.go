package utils

import "time"

// payloads
// payloads for append entries
type UserPayload struct {
	FirstName, LastName, HashedPassword, Email                                             string
	DateOfBirth                                                                            time.Time
	IdentificationNumber, IdentificationImageFront, IdentificationImageBack, PrevPW, NewPW string
	UserID                                                                                 int
	Action                                                                                 UserAction // create, update, delete
}

type AdminPayload struct {
	FirstName, LastName, HashedPassword, Email string
	AdminID, UserId                            int
	Action                                     AdminAction
}

type WalletOperationPayload struct {
	Wallet1, Wallet2 int
	Amount           int64
	Action           WalletAction
}
