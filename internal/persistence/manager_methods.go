package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rubberpipe/rubberpipe/internal/types"
)

func (s *ConfigManagerService) LogJobRecord(ctx context.Context, record types.BackupRecord) error {
	query := `
		INSERT INTO backups 
		(source, destination, filename, timestamp, status, error_msg)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(
		ctx,
		query,
		record.Source,
		record.Destination,
		record.Filename,
		record.Timestamp,
		record.Status,
		record.ErrorMsg,
	)
	if err != nil {
		return fmt.Errorf("failed to insert job record: %w", err)
	}
	return nil
}

func (s *ConfigManagerService) GetBackupRecord(ctx context.Context, backupID int) (types.BackupRecord, error) {
	var record types.BackupRecord
	query := `
		SELECT id, source, destination, filename, timestamp, status, error_msg 
		FROM backups WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, query, backupID)
	err := row.Scan(
		&record.ID,
		&record.Source,
		&record.Destination,
		&record.Filename,
		&record.Timestamp,
		&record.Status,
		&record.ErrorMsg,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return record, fmt.Errorf("backup record ID %d not found", backupID)
		}
		return record, fmt.Errorf("failed to scan backup record: %w", err)
	}
	return record, nil
}

func (s *ConfigManagerService) ListJobRecords(ctx context.Context) ([]types.BackupRecord, error) {
	query := `
		SELECT id, source, destination, filename, timestamp, status, error_msg 
		FROM backups ORDER BY timestamp DESC
	`
	return s.queryRecords(ctx, query)
}

func (s *ConfigManagerService) ListJobRecordsBySource(ctx context.Context, sourceName string) ([]types.BackupRecord, error) {
	query := `
		SELECT id, source, destination, filename, timestamp, status, error_msg 
		FROM backups WHERE source = ? ORDER BY timestamp DESC
	`
	return s.queryRecords(ctx, query, sourceName)
}

func (s *ConfigManagerService) queryRecords(ctx context.Context, query string, args ...interface{}) ([]types.BackupRecord, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	records := []types.BackupRecord{}
	for rows.Next() {
		var record types.BackupRecord
		if err := rows.Scan(&record.ID, &record.Source, &record.Destination, &record.Filename, &record.Timestamp, &record.Status, &record.ErrorMsg); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *ConfigManagerService) AddAdapterConfig(name, adapterType, configJSON string) error {
	query := `
        INSERT OR REPLACE INTO adapter_configs 
        (name, type, config_json) 
        VALUES (?, ?, ?)
    `
	_, err := s.db.Exec(query, name, adapterType, configJSON)
	if err != nil {
		return fmt.Errorf("failed to save adapter config: %w", err)
	}
	fmt.Printf("Adapter config '%s' added.\n", name)
	return nil
}

func (s *ConfigManagerService) RemoveAdapterConfig(name string) error {
	query := `DELETE FROM adapter_configs WHERE name = ?`
	result, err := s.db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to remove adapter config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("adapter config '%s' not found for removal", name)
	}
	fmt.Printf("Adapter config '%s' removed.\n", name)

	return nil
}

func (s *ConfigManagerService) GetAdapterConfigs() ([]types.AdapterConfig, error) {
	query := `SELECT name, type, config_json FROM adapter_configs ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query adapter configs: %w", err)
	}
	defer rows.Close()

	configs := []types.AdapterConfig{}

	for rows.Next() {
		var config types.AdapterConfig
		if err := rows.Scan(&config.Name, &config.Type, &config.ConfigJSON); err != nil {
			return nil, fmt.Errorf("failed to scan row into AdapterConfig: %w", err)
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in adapter configs: %w", err)
	}
	return configs, nil
}
