package analytics

import (
	compatibilitypostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/compatibility"
	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

type LegacyPublicKey = compatibilityapp.LegacyKey
type LegacyPublicMapping = compatibilityapp.LegacyMapping
type LegacyPublicDatabase = compatibilitypostgres.PublicDatabase

func NewLegacyPublicStore(
	database LegacyPublicDatabase,
) (compatibilityapp.PublicLegacyStore, error) {
	return compatibilitypostgres.NewPublicStore(database)
}
