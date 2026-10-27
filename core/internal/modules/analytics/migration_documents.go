package analytics

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.sql
var platformMigrationFiles embed.FS

//go:embed migrations/tenant/fresh/*.sql migrations/tenant/supersessions/fresh/*.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	platformDocuments, err := migration.CollectEmbedded(
		"analytics",
		"platform",
		"core/internal/modules/analytics",
		platformMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	tenantDocuments, err := migration.CollectEmbedded(
		"analytics",
		"tenant",
		"core/internal/modules/analytics",
		tenantMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	return append(platformDocuments, tenantDocuments...), nil
}
