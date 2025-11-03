package sources

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"github.com/rubberpipe/rubberpipe/internal/hub"
	"github.com/rubberpipe/rubberpipe/internal/types"
)

type PostgresAdapter struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	BackupDir string
}

type PostgresConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"dbname"`
	BackupDir string `json:"backup_dir"`
}

func NewPostgresAdapter(cfg PostgresConfig) *PostgresAdapter {
	return &PostgresAdapter{
		Host:      cfg.Host,
		Port:      cfg.Port,
		User:      cfg.User,
		Password:  cfg.Password,
		DBName:    cfg.DBName,
		BackupDir: cfg.BackupDir,
	}
}

func PostgresAdapterFactory(configJSON string) (hub.SourceAdapter, error) {
	var cfg PostgresConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid Postgres config JSON: %w", err)
	}
	return NewPostgresAdapter(cfg), nil
}

func init() {
	hub.RegisterSourceAdapter("postgres", PostgresAdapterFactory)
}

func (p *PostgresAdapter) Backup() (types.SourceArtifact, error) {
	if err := os.MkdirAll(p.BackupDir, os.ModePerm); err != nil {
		return types.SourceArtifact{}, fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(p.BackupDir, fmt.Sprintf("postgres_%s.dump", timestamp))

	cmd := exec.Command(
		"pg_dump",
		"-h", p.Host,
		"-p", fmt.Sprintf("%d", p.Port),
		"-U", p.User,
		"-F", "c",
		"-f", backupFile,
		p.DBName,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		os.Remove(backupFile)
		return types.SourceArtifact{}, fmt.Errorf("pg_dump failed: %w", err)
	}

	info, err := os.Stat(backupFile)
	if err != nil {
		return types.SourceArtifact{}, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return types.SourceArtifact{
		Path:          backupFile,
		FileSizeBytes: info.Size(),
		MimeType:      "application/x-postgres-dump",
		IsTemporary:   true,
		Metadata:      nil,
	}, nil
}

func (p *PostgresAdapter) Validate() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

func (p *PostgresAdapter) Restore(artifact types.SourceArtifact) error {
	filePath := artifact.Path

	cmd := exec.Command(
		"pg_restore",
		"-h", p.Host,
		"-p", fmt.Sprintf("%d", p.Port),
		"-U", p.User,
		"-d", p.DBName,
		"-c",
		filePath,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
