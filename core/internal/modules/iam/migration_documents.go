package iam

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
		"iam",
		"platform",
		"core/internal/modules/iam",
		platformMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	tenantDocuments, err := migration.CollectEmbedded(
		"iam",
		"tenant",
		"core/internal/modules/iam",
		tenantMigrationFiles,
	)
	if err != nil {
		return nil, err
	}
	return append(platformDocuments, tenantDocuments...), nil
}
