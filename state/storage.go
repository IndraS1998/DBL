package state

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"raft/utils"
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
// TODO add transaction status to each log entry
type LogEntry struct {
	Index          int `gorm:"primaryKey;autoIncrement"` // Log index
	Term           int32
	ReferenceTable utils.RefTable
	*utils.UserPayload
	*utils.AdminPayload
	*utils.WalletOperationPayload
}

// Initialize the Database and Auto-Migrate
func InitPersistentState(filePath string) (*PersistentState, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&MetaState{}, &LogEntry{})
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
func (ps *PersistentState) AppendLogEntry(term int32, refTable utils.RefTable, up *utils.UserPayload, ap *utils.AdminPayload, wp *utils.WalletOperationPayload) error {
	switch refTable {
	case utils.RefUser:
		if up == nil {
			return fmt.Errorf("user payload is required")
		}
		entry := LogEntry{Term: term, ReferenceTable: refTable, UserPayload: up, AdminPayload: nil, WalletOperationPayload: nil}
		return ps.DB.Create(&entry).Error
	case utils.RefAdmin:
		if ap == nil {
			return fmt.Errorf("admin payload is required")
		}
		entry := LogEntry{Term: term, ReferenceTable: refTable, UserPayload: nil, AdminPayload: ap, WalletOperationPayload: nil}
		return ps.DB.Create(&entry).Error
	case utils.RefWallet:
		if wp == nil {
			return fmt.Errorf("wallet payload is required")
		}
		entry := LogEntry{Term: term, ReferenceTable: refTable, UserPayload: nil, AdminPayload: nil, WalletOperationPayload: wp}
		return ps.DB.Create(&entry).Error
	default:
		return fmt.Errorf("invalid reference table: %s", refTable)
	}
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

func (ps *PersistentState) GetCommandsFromIndex(startIndex int) ([]string, error) {
	var commands []string
	err := ps.DB.Model(&LogEntry{}).
		Where("`index` >= ?", startIndex).
		Order("`index` asc").
		Pluck("cmd", &commands).Error
	if err != nil {
		return nil, err
	}
	return commands, nil
}
