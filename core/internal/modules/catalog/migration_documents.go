package catalog

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
		"catalog",
		"platform",
		"core/internal/modules/catalog",
		platformMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	tenantDocuments, err := migration.CollectEmbedded(
		"catalog",
		"tenant",
		"core/internal/modules/catalog",
		tenantMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	return append(platformDocuments, tenantDocuments...), nil
}
