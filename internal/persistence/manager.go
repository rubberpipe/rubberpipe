package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rubberpipe/rubberpipe/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

type ConfigManager interface {
	LogJobRecord(ctx context.Context, record types.BackupRecord) error

	GetBackupRecord(ctx context.Context, backupID int) (types.BackupRecord, error)
	ListJobRecords(ctx context.Context) ([]types.BackupRecord, error)
	ListJobRecordsBySource(ctx context.Context, sourceName string) ([]types.BackupRecord, error)

	AddAdapterConfig(name, adapterType, configJSON string) error
	RemoveAdapterConfig(name string) error
	GetAdapterConfigs() ([]types.AdapterConfig, error)
}

type ConfigManagerService struct {
	db *sql.DB
}

func NewConfigManagerService(db *sql.DB) (ConfigManager, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if err := setupDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to set up database schema: %w", err)
	}

	return &ConfigManagerService{db: db}, nil
}
func setupDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY,
			source TEXT,
			destination TEXT,
			filename TEXT,
			timestamp DATETIME,
			status TEXT,
			error_msg TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Adapter Configs Table Schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS adapter_configs (
			name TEXT PRIMARY KEY,    
			type TEXT,                
			config_json TEXT          
		)
	`)
	if err != nil {
		return err
	}

	return nil
}
