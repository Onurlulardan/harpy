package platformauthority

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.fresh.sql
var platformMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"platformauthority", "platform", "core/internal/modules/platformauthority",
		platformMigrationFiles,
	)
}
