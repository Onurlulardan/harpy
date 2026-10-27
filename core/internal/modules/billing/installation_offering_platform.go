package billing

import (
	"context"

	operatorbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/operatorbinding"
	"github.com/google/uuid"
)

type InstallationOfferingPlatformReader struct {
	database operatorbindingpostgres.OfferingDatabase
}

func NewInstallationOfferingPlatformReader(
	pool operatorbindingpostgres.Pool,
) (*InstallationOfferingPlatformReader, error) {
	return operatorbindingpostgres.WithInstallationOffering(pool,
		func(database operatorbindingpostgres.OfferingDatabase) (
			*InstallationOfferingPlatformReader, error,
		) {
			return &InstallationOfferingPlatformReader{database: database}, nil
		})
}

func (reader *InstallationOfferingPlatformReader) FetchInstallationOffering(
	ctx context.Context,
	environmentID uuid.UUID,
	mode string,
) (InstallationOfferingPin, error) {
	if reader == nil || reader.database == nil {
		return InstallationOfferingPin{}, ErrInstallationOfferingUnavailable
	}
	value, err := reader.database.FetchInstallationOffering(ctx, environmentID, mode)
	return InstallationOfferingPin{
		VersionID: value.VersionID, Revision: value.Revision,
		SourceSequence: value.SourceSequence, SHA256: value.SHA256,
		SourceSHA256: value.SourceSHA256, EffectiveAt: value.EffectiveAt,
		SourcePresent: value.SourcePresent, Mode: mode,
	}, err
}

var _ InstallationOfferingReader = (*InstallationOfferingPlatformReader)(nil)
