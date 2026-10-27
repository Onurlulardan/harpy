package support

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.fresh.sql
var platformMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"support", "platform", "core/internal/modules/support", platformMigrationFiles,
	)
}
