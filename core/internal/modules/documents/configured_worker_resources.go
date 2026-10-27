package documents

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/adoptionartifact"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type configuredWorkerResources struct {
	platformArtifacts
	legacy                *adoptionartifact.S3Source
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
	legacySettings, err := settings.LegacySource()
	if err != nil {
		return configuredWorkerResources{}, err
	}
	client, bucket, owner, err := openPlatformS3Client(ctx, legacySettings)
	if err != nil {
		return configuredWorkerResources{}, err
	}
	legacy, err := adoptionartifact.NewVerifiedS3Source(ctx, client, bucket, owner)
	if err != nil {
		return configuredWorkerResources{}, fmt.Errorf(
			"verify Documents legacy S3 source: %w", err)
	}
	ttl, err := settings.ArchiveReservationTTL()
	if err != nil {
		return configuredWorkerResources{}, err
	}
	return configuredWorkerResources{platformArtifacts: artifacts, legacy: legacy,
		archiveReservationTTL: ttl}, nil
}
