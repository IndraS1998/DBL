package state

import (
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
type LogEntry struct {
	Index int `gorm:"primaryKey;autoIncrement"` // Log index
	Term  int32
	Cmd   string
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
func (ps *PersistentState) AppendLogEntry(term int32, cmd string) error {
	entry := LogEntry{Term: term, Cmd: cmd}
	return ps.DB.Create(&entry).Error
}

func (ps *PersistentState) GetLogEntry(index int) (*LogEntry, error) {
	var entry LogEntry
	err := ps.DB.First(&entry, index).Error
	return &entry, err
}

func (ps *PersistentState) GetLastLogIndex() (int, error) {
	var entry LogEntry
	err := ps.DB.Order("index desc").First(&entry).Error
	return entry.Index, err
}

func (ps *PersistentState) DeleteLogEntriesFrom(index int) error {
	return ps.DB.Where("index >= ?", index).Delete(&LogEntry{}).Error
}

func (ps *PersistentState) GetLogLengh() (int64, error) {
	var count int64
	err := ps.DB.Model(&LogEntry{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
