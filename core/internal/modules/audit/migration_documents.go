package audit

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/tenant/fresh/*.fresh.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"audit", "tenant", "core/internal/modules/audit", tenantMigrationFiles,
	)
}
