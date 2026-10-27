package systemresourcelayout

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/tenant/fresh/*.fresh.sql
//go:embed migrations/tenant/supersessions/fresh/*.fresh.sql
//go:embed migrations/tenant/supersessions/v2/fresh/*.fresh.sql
//go:embed migrations/tenant/supersessions/v3/fresh/*.fresh.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"system_resource_layout",
		"tenant",
		"core/internal/modules/system_resource_layout",
		tenantMigrationFiles,
	)
}
