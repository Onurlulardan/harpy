package migrator

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/fresh/*.sql
var migrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"platform_migrator",
		"platform",
		"core/internal/platform/migrator",
		migrationFiles,
	)
}
