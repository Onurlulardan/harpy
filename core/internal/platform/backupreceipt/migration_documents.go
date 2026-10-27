package backupreceipt

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/fresh/*.sql
var migrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	return migration.CollectEmbedded(
		"platform_backupreceipt",
		"platform",
		"core/internal/platform/backupreceipt",
		migrationFiles,
	)
}
