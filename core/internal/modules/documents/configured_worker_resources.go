package documents

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type configuredWorkerResources struct {
	platformArtifacts
	archiveReservationTTL time.Duration
}

func openConfiguredWorkerResources(ctx context.Context,
	settings config.DocumentsWorker,
) (configuredWorkerResources, error) {
	artifactSettings, err := settings.Artifacts()
	if err != nil {
		return configuredWorkerResources{}, err
	}
	artifacts, err := openPlatformArtifacts(ctx, artifactSettings)
	if err != nil {
		return configuredWorkerResources{}, err
	}
	ttl, err := settings.ArchiveReservationTTL()
	if err != nil {
		return configuredWorkerResources{}, err
	}
	return configuredWorkerResources{platformArtifacts: artifacts,
		archiveReservationTTL: ttl}, nil
}
