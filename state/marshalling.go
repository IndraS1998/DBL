package state

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	pb "raft/raft"
	"raft/utils"
)

// proto log entry -> gorm log entry
func ProtoToLogEntry(entry *pb.LogEntry, tableRef string) (utils.Payload, error) {
	switch tableRef {
	case string(utils.RefUser):
		userPayload, ok := entry.Payload.(*pb.LogEntry_UserPayload)
		if ok {
			return utils.UserPayload{
				FirstName:                userPayload.UserPayload.FirstName,
				LastName:                 userPayload.UserPayload.LastName,
				HashedPassword:           userPayload.UserPayload.HashedPassword,
				Email:                    userPayload.UserPayload.Email,
				DateOfBirth:              userPayload.UserPayload.DateOfBirth.AsTime(),
				IdentificationNumber:     userPayload.UserPayload.IdentificationNumber,
				IdentificationImageFront: userPayload.UserPayload.IdentificationImageFront,
				IdentificationImageBack:  userPayload.UserPayload.IdentificationImageBack,
				PrevPW:                   userPayload.UserPayload.PrevPW,
				NewPW:                    userPayload.UserPayload.NewPW,
				UserID:                   int(userPayload.UserPayload.UserID),
				Action:                   utils.UserAction(userPayload.UserPayload.Action),
			}, nil
		} else {
			return utils.UserPayload{}, fmt.Errorf("failed to cast payload to UserPayload")
		}
	case string(utils.RefAdmin):
		adminPayload, ok := entry.Payload.(*pb.LogEntry_AdminPayload)
		if ok {
			return utils.AdminPayload{
				FirstName:      adminPayload.AdminPayload.FirstName,
				LastName:       adminPayload.AdminPayload.LastName,
				HashedPassword: adminPayload.AdminPayload.HashedPassword,
				Email:          adminPayload.AdminPayload.Email,
				AdminID:        int(adminPayload.AdminPayload.AdminID),
				UserId:         int(adminPayload.AdminPayload.UserId),
				Action:         utils.AdminAction(adminPayload.AdminPayload.Action),
			}, nil
		} else {
			return utils.AdminPayload{}, fmt.Errorf("failed to cast payload to AdminPayload")
		}
	case string(utils.RefWallet):
		walletPayload, ok := entry.Payload.(*pb.LogEntry_WalletOperationPayload)
		if ok {
			return utils.WalletOperationPayload{
				Wallet1: int(walletPayload.WalletOperationPayload.Wallet1),
				Wallet2: int(walletPayload.WalletOperationPayload.Wallet2),
				Amount:  walletPayload.WalletOperationPayload.Amount,
				Action:  utils.WalletAction(walletPayload.WalletOperationPayload.Action),
			}, nil
		} else {
			return utils.WalletOperationPayload{}, fmt.Errorf("failed to cast payload to Wallet operation")
		}
	default:
		return utils.UserPayload{}, fmt.Errorf("unsopported table reference:%s", tableRef)
	}
}

// Gorm log Entry -> proto Lof entry
func ToProtoLogEntry(entry LogEntry, db *gorm.DB) (*pb.LogEntry, error) {
	refTable := entry.ReferenceTable

	switch refTable {
	case utils.RefUser:
		var payload UserPayload
		if err := db.First(&payload, entry.PayloadID).Error; err != nil {
			return nil, fmt.Errorf("failed to load user payload: %w", err)
		}
		return &pb.LogEntry{
			Index:          int64(entry.Index),
			Term:           entry.Term,
			ReferenceTable: string(refTable),
			Payload: &pb.LogEntry_UserPayload{
				UserPayload: &pb.UserPayload{
					FirstName:                *payload.FirstName,
					LastName:                 *payload.LastName,
					HashedPassword:           *payload.HashedPassword,
					Email:                    *payload.Email,
					DateOfBirth:              timestamppb.New(*payload.DateOfBirth),
					IdentificationNumber:     *payload.IdentificationNumber,
					IdentificationImageFront: *payload.IdentificationImageFront,
					IdentificationImageBack:  *payload.IdentificationImageBack,
					PrevPW:                   *payload.PrevPW,
					NewPW:                    *payload.NewPW,
					UserID:                   int64(*payload.UserID),
					Action:                   string(payload.Action),
				},
			},
		}, nil

	case utils.RefAdmin:
		var payload AdminPayload
		if err := db.First(&payload, entry.PayloadID).Error; err != nil {
			return nil, fmt.Errorf("failed to load admin payload: %w", err)
		}
		if payload.FirstName == nil || payload.LastName == nil || payload.HashedPassword == nil || payload.Email == nil {
			fmt.Println("what i just edited")
			fmt.Println(entry.Term, string(refTable), payload.FirstName, payload.LastName,
				payload.HashedPassword, payload.Action, payload.AdminID, payload.UserId)
			return nil, fmt.Errorf("nil field")
		}
		return &pb.LogEntry{
			Index:          int64(entry.Index),
			Term:           entry.Term,
			ReferenceTable: string(refTable),
			Payload: &pb.LogEntry_AdminPayload{
				AdminPayload: &pb.AdminPayload{
					FirstName:      *payload.FirstName,
					LastName:       *payload.LastName,
					HashedPassword: *payload.HashedPassword,
					Email:          *payload.Email,
					AdminID:        int64(*payload.AdminID),
					UserId:         int64(*payload.UserId),
					Action:         string(payload.Action),
				},
			},
		}, nil

	case utils.RefWallet:
		var payload WalletOperationPayload
		if err := db.First(&payload, entry.PayloadID).Error; err != nil {
			return nil, fmt.Errorf("failed to load wallet payload: %w", err)
		}
		return &pb.LogEntry{
			Index:          int64(entry.Index),
			Term:           entry.Term,
			ReferenceTable: string(refTable),
			Payload: &pb.LogEntry_WalletOperationPayload{
				WalletOperationPayload: &pb.WalletOperationPayload{
					Wallet1: int64(payload.Wallet1),
					Wallet2: int64(*payload.Wallet2),
					Amount:  payload.Amount,
					Action:  string(payload.Action),
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported reference table: %s", refTable)
	}
}
