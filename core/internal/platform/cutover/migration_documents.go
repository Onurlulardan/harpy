package cutover

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/fresh/*.sql migrations/supersessions/fresh/*.sql
var migrationFiles embed.FS

//go:embed migration-supersessions.json
var migrationSupersessionCatalog []byte

func MigrationDocuments() ([]migration.Document, error) {
	documents, err := migration.CollectEmbedded(
		"platform_cutover",
		"platform",
		"core/internal/platform/cutover",
		migrationFiles,
	)
	if err != nil {
		return nil, err
	}
	return selectMigrationSupersessions(documents)
}
