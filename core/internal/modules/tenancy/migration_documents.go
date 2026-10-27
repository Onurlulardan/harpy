package tenancy

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.sql migrations/platform/supersessions/fresh/*.sql
var platformMigrationFiles embed.FS

//go:embed migrations/tenant/fresh/*.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	platformDocuments, err := migration.CollectEmbedded(
		"tenancy",
		"platform",
		"core/internal/modules/tenancy",
		platformMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	tenantDocuments, err := migration.CollectEmbedded(
		"tenancy",
		"tenant",
		"core/internal/modules/tenancy",
		tenantMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	return append(platformDocuments, tenantDocuments...), nil
}
