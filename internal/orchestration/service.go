package orchestration

import (
	"context"

	"github.com/rubberpipe/rubberpipe/internal/hub"
	"github.com/rubberpipe/rubberpipe/internal/persistence"
	"github.com/rubberpipe/rubberpipe/internal/types"
)

/*
The OrchestratorService is the centralized abstraction for the entire application.

OrchestratorService implements Orchestrator interface. The struct will then implement
the methods by calling the Hub and ConfigManager counterpart.
*/
type OrchestratorService struct {
	Execution     hub.Hub
	ConfigManager persistence.ConfigManager
}

func NewOrchestratorService(exec hub.Hub, config persistence.ConfigManager) (Orchestrator, error) {
	return &OrchestratorService{
		Execution:     exec,
		ConfigManager: config,
	}, nil
}

func (o *OrchestratorService) AddAdapterConfig(name, adapterType, configJSON string) error {
	return o.ConfigManager.AddAdapterConfig(name, adapterType, configJSON)
}

func (o *OrchestratorService) RemoveAdapterConfig(name string) error {
	err := o.ConfigManager.RemoveAdapterConfig(name)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrchestratorService) ListAdapterConfigs() ([]types.AdapterConfig, error) {
	return o.ConfigManager.GetAdapterConfigs()
}

func (o *OrchestratorService) ListBackupHistory(ctx context.Context) ([]types.BackupRecord, error) {
	return o.ConfigManager.ListJobRecords(ctx)
}

func (o *OrchestratorService) Backup(ctx context.Context, sourceName string, destName string) (types.ArtifactID, error) {
	return o.Execution.DoBackupJob(ctx, sourceName, destName)
}

func (o *OrchestratorService) Restore(ctx context.Context, backupID int) error {
	return nil
}
