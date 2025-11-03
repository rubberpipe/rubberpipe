package types

import "time"

type ArtifactID struct {
	Identifier      string
	DestinationType string
}

type SourceArtifact struct {
	Path          string // The path to the created artifact (e.g., /tmp/db.dump)
	MimeType      string // The type of artifact (e.g., "application/x-postgres-dump")
	FileSizeBytes int64
	IsTemporary   bool              // Flag to indicate if this artifact needs to be cleaned up after storage
	Metadata      map[string]string // Optional key/value pair for adapter-specific data
}

type BackupRecord struct {
	ID          int       `json:"id"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Filename    string    `json:"filename"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	ErrorMsg    string    `json:"error_msg"`
}

type AdapterConfig struct {
	Name       string `json:"name"`        // The unique identifier/handle (e.g., "postgres_main")
	Type       string `json:"type"`        // The technology (e.g., "postgres" or "local")
	ConfigJSON string `json:"config_json"` // The JSON string of adapter-specific settings
}
