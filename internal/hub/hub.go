package hub

import (
	"context"
	"fmt"
	"os"

	"github.com/rubberpipe/rubberpipe/internal/types"
)

type SourceAdapter interface {
	Backup() (types.SourceArtifact, error)
	Validate() error
	Restore(types.SourceArtifact) error
}

type DestinationAdapter interface {
	Store(rawPath string) (types.ArtifactID, error)
	Retrieve(id types.ArtifactID) (string, error)
}

type Hub interface {
	DoBackupJob(ctx context.Context, sourceName, destName string) (types.ArtifactID, error)
	DoRestoreJob(ctx context.Context, srcAdapterName string, artifact types.ArtifactID) error
}

type HubService struct {
	Sources      map[string]SourceAdapter
	Destinations map[string]DestinationAdapter
}

func NewHub(configs []types.AdapterConfig) (Hub, error) {
	hubService := &HubService{
		Sources:      make(map[string]SourceAdapter),
		Destinations: make(map[string]DestinationAdapter),
	}

	for _, cfg := range configs {
		if factory, ok := sourceFactories[cfg.Type]; ok {
			adapter, err := factory(cfg.ConfigJSON)
			if err != nil {
				return nil, fmt.Errorf("hub: failed to init source adapter %s: %w", cfg.Name, err)
			}
			hubService.Sources[cfg.Name] = adapter
		} else if factory, ok := destinationFactories[cfg.Type]; ok {
			adapter, err := factory(cfg.ConfigJSON)
			if err != nil {
				return nil, fmt.Errorf("hub: failed to init destination adapter %s: %w", cfg.Name, err)
			}
			hubService.Destinations[cfg.Name] = adapter
		} else {
			return nil, fmt.Errorf("hub: unknown adapter type: %s", cfg.Type)
		}
	}

	return hubService, nil
}

func (h *HubService) DoBackupJob(ctx context.Context, sourceName, destName string) (types.ArtifactID, error) {
	src, ok := h.Sources[sourceName]
	if !ok {
		return types.ArtifactID{}, fmt.Errorf("source adapter %s not found", sourceName)
	}
	dest, ok := h.Destinations[destName]
	if !ok {
		return types.ArtifactID{}, fmt.Errorf("destination adapter %s not found", destName)
	}

	// 1. Source Stage: Get the raw, consistent data dump.
	artifact, err := src.Backup()
	if err != nil {
		return types.ArtifactID{}, fmt.Errorf("source dump failed: %w", err)
	}
	if artifact.IsTemporary {
		defer os.RemoveAll(artifact.Path)
	}

	// TODO: Restic Middleware

	// Destination Stage: Store the raw artifact file directly.
	artifactID, err := dest.Store(artifact.Path)
	if err != nil {
		return types.ArtifactID{}, fmt.Errorf("destination store failed: %w", err)
	}

	return artifactID, nil
}

// DoRestoreJob: The central restoration pipeline method.
func (h *HubService) DoRestoreJob(ctx context.Context, srcAdapterName string, artifactID types.ArtifactID) error {
	// NOTE: The Orchestrator handles the persistence lookup to get the ArtifactID.
	// This method assumes the artifact has been retrieved locally (Decoupling Assumption).

	src, ok := h.Sources[srcAdapterName]
	if !ok {
		return fmt.Errorf("source adapter %s not found", srcAdapterName)
	}

	// For demonstration, we assume we have the original artifact structure (which needs to be stored
	// in the BackupRecord in a real app) and create a temp object.
	// This is where we would call Destination.Retrieve and get the local file path.
	tempLocalPath, err := h.retrieveArtifactLocally(ctx, artifactID)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempLocalPath)

	// Create a minimal SourceArtifact pointing to the retrieved, raw file.
	rawArtifact := types.SourceArtifact{
		Path:        tempLocalPath,
		IsTemporary: true,
	}

	// 2. Source Stage Reverse: Execute the restore using the retrieved artifact data.
	if err := src.Restore(rawArtifact); err != nil {
		return fmt.Errorf("source restore failed: %w", err)
	}

	return nil
}

// Helper to simulate retrieval and decryption (currently just retrieval)
func (h *HubService) retrieveArtifactLocally(ctx context.Context, id types.ArtifactID) (string, error) {
	dest, ok := h.Destinations[id.DestinationType]
	if !ok {
		return "", fmt.Errorf("destination adapter type %s not found for retrieval", id.DestinationType)
	}

	// Call the adapter's retrieve method using the ArtifactID
	localFilePath, err := dest.Retrieve(id)
	if err != nil {
		return "", fmt.Errorf("destination retrieval failed: %w", err)
	}
	return localFilePath, nil
}
