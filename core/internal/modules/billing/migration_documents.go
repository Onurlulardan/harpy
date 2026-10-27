package billing

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.sql
var platformMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"billing",
		"platform",
		"core/internal/modules/billing",
		platformMigrationFiles,
	)
}
