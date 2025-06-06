package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	PollID                                                                                 string
	Action                                                                                 UserAction
	Term                                                                                   int32
}

func (up UserPayload) GetRefTable() RefTable {
	return RefUser
}

type AdminPayload struct {
	FirstName, LastName, HashedPassword, Email string
	AdminID, UserId                            int
	PollID                                     string
	Action                                     AdminAction
	Term                                       int32
}

func (ap AdminPayload) GetRefTable() RefTable {
	return RefAdmin
}

type WalletOperationPayload struct {
	Wallet1, Wallet2 int
	Amount           int64
	PollID           string
	Action           WalletAction
	Term             int32
}

func (wp WalletOperationPayload) GetRefTable() RefTable {
	return RefWallet
}

type PayloadWrapper struct {
	Ref  string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func WrapPayload(p Payload) (*PayloadWrapper, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return &PayloadWrapper{
		Ref:  string(p.GetRefTable()),
		Data: data,
	}, nil
}

func UnwrapPayload(wrappedStr string) (Payload, error) {
	var wrapper PayloadWrapper
	if err := json.Unmarshal([]byte(wrappedStr), &wrapper); err != nil {
		return nil, err
	}

	switch wrapper.Ref {
	case string(RefUser):
		var p UserPayload
		err := json.Unmarshal(wrapper.Data, &p)
		return p, err
	case string(RefAdmin):
		var p AdminPayload
		err := json.Unmarshal(wrapper.Data, &p)
		return p, err
	case string(RefWallet):
		var p WalletOperationPayload
		err := json.Unmarshal(wrapper.Data, &p)
		return p, err
	default:
		return nil, fmt.Errorf("unknown payload type: %s", wrapper.Ref)
	}
}
