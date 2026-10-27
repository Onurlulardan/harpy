package work

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/tenant/fresh/*.fresh.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"work", "tenant", "core/internal/modules/work", tenantMigrationFiles,
	)
}
