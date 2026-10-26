package operatorbinding

import (
	"context"
	_ "embed"
	"time"

	"github.com/google/uuid"
)

//go:embed sql/platform/installation_offering.sql
var installationOfferingSQL string

type InstallationOffering struct {
	VersionID                uuid.UUID
	Revision, SourceSequence int64
	SHA256, SourceSHA256     string
	EffectiveAt              time.Time
	SourcePresent            bool
}

type OfferingDatabase interface {
	FetchInstallationOffering(context.Context, uuid.UUID, string) (InstallationOffering, error)
}

func (value offeringDatabase) FetchInstallationOffering(
	ctx context.Context,
	environmentID uuid.UUID,
	mode string,
) (InstallationOffering, error) {
	var result InstallationOffering
	var sourceSequence *int64
	var sourceSHA *string
	err := value.bound.QueryRow(ctx, installationOfferingSQL, environmentID, mode).Scan(
		&result.VersionID, &result.Revision, &result.SHA256,
		&result.EffectiveAt, &sourceSequence, &sourceSHA,
	)
	if err != nil {
		return InstallationOffering{}, err
	}
	if sourceSequence != nil && sourceSHA != nil {
		result.SourcePresent = true
		result.SourceSequence, result.SourceSHA256 = *sourceSequence, *sourceSHA
	} else if sourceSequence != nil || sourceSHA != nil {
		return InstallationOffering{}, errInvalidInstallationOffering
	}
	result.EffectiveAt = result.EffectiveAt.UTC()
	return result, nil
}
