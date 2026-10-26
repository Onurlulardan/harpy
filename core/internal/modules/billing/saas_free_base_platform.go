package billing

import (
	"context"

	operatorbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/operatorbinding"
)

type SaaSFreeBasePlatformReader struct {
	database operatorbindingpostgres.SaaSBaseDatabase
}

func NewSaaSFreeBasePlatformReader(
	pool operatorbindingpostgres.Pool,
) (*SaaSFreeBasePlatformReader, error) {
	return operatorbindingpostgres.WithSaaSBase(pool,
		func(database operatorbindingpostgres.SaaSBaseDatabase) (
			*SaaSFreeBasePlatformReader, error,
		) {
			return &SaaSFreeBasePlatformReader{database: database}, nil
		})
}

func (reader *SaaSFreeBasePlatformReader) FetchSaaSFreeBase(
	ctx context.Context,
) (SaaSFreeBasePin, error) {
	if reader == nil || reader.database == nil {
		return SaaSFreeBasePin{}, ErrSaaSFreeBaseUnavailable
	}
	value, err := reader.database.FetchSaaSBase(ctx)
	return SaaSFreeBasePin{
		VersionID: value.VersionID, Revision: value.Revision,
		SHA256: value.SHA256, EffectiveAt: value.EffectiveAt,
	}, err
}

var _ SaaSFreeBaseReader = (*SaaSFreeBasePlatformReader)(nil)
