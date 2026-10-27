package notifications

import (
	"embed"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

//go:embed migrations/platform/fresh/*.fresh.sql
var platformMigrationFiles embed.FS

//go:embed migrations/tenant/fresh/*.fresh.sql
var tenantMigrationFiles embed.FS

func MigrationDocuments() ([]migration.Document, error) {
	platform, err := migration.CollectEmbedded(
		"notifications", "platform", "core/internal/modules/notifications", platformMigrationFiles)
	if err != nil {
		return nil, err
	}
	tenant, err := migration.CollectEmbedded(
		"notifications", "tenant", "core/internal/modules/notifications", tenantMigrationFiles)
	if err != nil {
		return nil, err
	}
	return append(platform, tenant...), nil
}
