package state

import (
	"fmt"
	"raft/utils"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PersistentState struct {
	DB *gorm.DB
}

// This table stores the current term and who got the vote in that term
type MetaState struct {
	ID          int `gorm:"primaryKey"` // Always 1, singleton pattern
	CurrentTerm int32
	VotedFor    string
}

// Log entries are stored in their own table
type UserPayload struct {
	ID                                                                                     uint `gorm:"primaryKey"`
	FirstName, LastName, HashedPassword, Email                                             *string
	DateOfBirth                                                                            *time.Time
	IdentificationNumber, IdentificationImageFront, IdentificationImageBack, PrevPW, NewPW *string
	UserID                                                                                 *int
	Action                                                                                 utils.UserAction
}

type AdminPayload struct {
	ID                                         uint `gorm:"primaryKey"`
	FirstName, LastName, HashedPassword, Email *string
	AdminID, UserId                            *int
	Action                                     utils.AdminAction
}

type WalletOperationPayload struct {
	ID      uint `gorm:"primaryKey"`
	Wallet1 int
	Wallet2 *int
	Amount  int64
	Action  utils.WalletAction
}
type LogEntry struct {
	Index          int `gorm:"primaryKey;autoIncrement"` // Log index
	Term           int32
	ReferenceTable utils.RefTable
	Status         utils.TransactionStatus `gorm:"default:'pending'"`
	PayloadID      uint
}

// Initialize the Database and Auto-Migrate
func InitPersistentState(filePath string) (*PersistentState, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&MetaState{}, &UserPayload{}, &WalletOperationPayload{}, &AdminPayload{}, &LogEntry{})
	if err != nil {
		return nil, err
	}

	// Initialize singleton meta state if not present
	var meta MetaState
	if err := db.First(&meta, 1).Error; err != nil {
		// Create default if not found
		db.Create(&MetaState{ID: 1, CurrentTerm: 0, VotedFor: ""})
	}

	return &PersistentState{DB: db}, nil
}

// Read/Write Methods for MetaState
func (ps *PersistentState) SetCurrentTerm(term int32) error {
	return ps.DB.Model(&MetaState{}).Where("id = ?", 1).Update("current_term", term).Error
}

func (ps *PersistentState) GetCurrentTerm() (int32, error) {
	var meta MetaState
	err := ps.DB.First(&meta, 1).Error
	return meta.CurrentTerm, err
}

func (ps *PersistentState) SetVotedFor(candidateID string) error {
	return ps.DB.Model(&MetaState{}).Where("id = ?", 1).Update("voted_for", candidateID).Error
}

func (ps *PersistentState) GetVotedFor() (string, error) {
	var meta MetaState
	err := ps.DB.First(&meta, 1).Error
	return meta.VotedFor, err
}

// Managing Log Entries (Append, Read, Delete)
func (ps *PersistentState) AppendLogEntry(term int32, p utils.Payload) error {
	refTable := p.GetRefTable()
	return ps.DB.Transaction(func(tx *gorm.DB) error {
		switch refTable {
		case utils.RefUser:
			payload, ok := p.(utils.UserPayload)
			if ok {
				userPayload := UserPayload{
					FirstName:                &payload.FirstName,
					LastName:                 &payload.LastName,
					HashedPassword:           &payload.HashedPassword,
					Email:                    &payload.Email,
					DateOfBirth:              &payload.DateOfBirth,
					IdentificationNumber:     &payload.IdentificationNumber,
					IdentificationImageFront: &payload.IdentificationImageFront,
					IdentificationImageBack:  &payload.IdentificationImageBack,
					PrevPW:                   &payload.PrevPW,
					NewPW:                    &payload.NewPW,
					UserID:                   &payload.UserID,
					Action:                   payload.Action,
				}
				if err := tx.Create(&userPayload).Error; err != nil {
					return fmt.Errorf("failed to created the associated user payload:%w", err)
				}
				logEntry := LogEntry{
					Term: term, ReferenceTable: refTable, PayloadID: userPayload.ID,
				}
				if err := tx.Create(&logEntry).Error; err != nil {
					return fmt.Errorf("failed to create the log entry:%w", err)
				}
				return nil
			} else {
				return fmt.Errorf("failed to cast payload to UserPayload")
			}
		case utils.RefAdmin:
			payload, ok := p.(utils.AdminPayload)
			if ok {
				adminPayload := AdminPayload{
					FirstName: &payload.FirstName,
					LastName:  &payload.LastName,
					Email:     &payload.Email,
					AdminID:   &payload.AdminID,
					UserId:    &payload.UserId,
					Action:    payload.Action,
				}
				if err := tx.Create(&adminPayload).Error; err != nil {
					return fmt.Errorf("failed to create admin payload: %w", err)
				}
				logEntry := LogEntry{
					Term: term, ReferenceTable: refTable, PayloadID: adminPayload.ID,
				}
				if err := tx.Create(&logEntry).Error; err != nil {
					return fmt.Errorf("failed to create log entry for admin payload:%w", err)
				}
				return nil
			} else {
				return fmt.Errorf("failed to cast payload as AdminPayload")
			}
		case utils.RefWallet:
			payload, ok := p.(utils.WalletOperationPayload)
			if ok {
				walletPayload := WalletOperationPayload{
					Wallet1: payload.Wallet1,
					Wallet2: &payload.Wallet2,
					Amount:  payload.Amount,
					Action:  payload.Action,
				}
				if err := tx.Create(&walletPayload).Error; err != nil {
					return fmt.Errorf("failed to create wallet payload: %w", err)
				}
				logEntry := LogEntry{
					Term: term, ReferenceTable: refTable, PayloadID: walletPayload.ID,
				}
				if err := tx.Create(&logEntry).Error; err != nil {
					return fmt.Errorf("failed to create log entry for wallet payload:%w", err)
				}
				return nil
			} else {
				return fmt.Errorf("failed to cast payload as wallet operation")
			}
		default:
			return fmt.Errorf("unsupported operation: %s", refTable)
		}
	})
}

func (ps *PersistentState) GetLogEntry(index int) (*LogEntry, error) {
	var entry LogEntry
	err := ps.DB.First(&entry, index).Error
	return &entry, err
}

func (ps *PersistentState) GetAllLogEntries() ([]LogEntry, error) {
	var entries []LogEntry
	err := ps.DB.Order("`index` asc").Find(&entries).Error
	return entries, err
}

func (ps *PersistentState) GetLastLogEntry() (LogEntry, error) {
	var entry LogEntry
	err := ps.DB.Order("`index` desc").First(&entry).Error
	return entry, err
}

func (ps *PersistentState) DeleteLogEntriesFrom(index int) error {
	return ps.DB.Where("`index` >= ?", index).Delete(&LogEntry{}).Error
}

func (ps *PersistentState) GetLogLength() (int64, error) {
	var count int64
	err := ps.DB.Model(&LogEntry{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ps *PersistentState) GetCommandsFromIndex(startIndex int) ([]LogEntry, error) {
	var entries []LogEntry

	err := ps.DB.Model(&LogEntry{}).
		Where("`index` >= ?", startIndex).
		Order("`index` asc").Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("error getting entries: %w", err)
	}
	return entries, nil
}
func (ps *PersistentState) GetEntriesForCommit(lastApplied, commitIndex int) ([]LogEntry, error) {
	var entries []LogEntry
	err := ps.DB.Model(&LogEntry{}).
		Where("`index` > ? AND `index` <= ?", lastApplied, commitIndex).
		Order("`index` asc").Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("error getting entries: %w", err)
	}
	return entries, nil
}
