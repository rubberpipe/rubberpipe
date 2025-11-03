package orchestration

import (
	"context"

	"github.com/rubberpipe/rubberpipe/internal/types"
)

type Orchestrator interface {
	Backup(ctx context.Context, sourceName, destName string) (types.ArtifactID, error)
	Restore(ctx context.Context, backupID int) error

	AddAdapterConfig(name, adapterType, configJSON string) error
	RemoveAdapterConfig(name string) error
	ListAdapterConfigs() ([]types.AdapterConfig, error)

	ListBackupHistory(ctx context.Context) ([]types.BackupRecord, error)
}
