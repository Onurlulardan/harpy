package migrator

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/modules/support"
	systemresourcelayout "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/work"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/backupreceipt"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

type migrationCollection struct {
	name string
	load func() ([]migration.Document, error)
}

func FreshMigrationDocuments() ([]migration.Document, error) {
	return deployMigrationDocuments()
}

func deployMigrationDocuments() ([]migration.Document, error) {
	var documents []migration.Document
	for _, collection := range platformMigrationCollections() {
		owned, err := collection.load()
		if err != nil {
			return nil, fmt.Errorf("load %s migrations: %w", collection.name, err)
		}
		documents = append(documents, owned...)
	}
	selected, err := migration.SelectFreshDocuments(documents)
	if err != nil {
		return nil, fmt.Errorf("select fresh migrations: %w", err)
	}
	return selected, nil
}

func platformMigrationCollections() []migrationCollection {
	return []migrationCollection{
		{"platform migrator", platformmigrator.MigrationDocuments},
		{"platform cutover", cutover.MigrationDocuments},
		{"tenancy", tenancy.MigrationDocuments},
		{"platform authority", platformauthority.MigrationDocuments},
		{"platform operations", platformops.MigrationDocuments},
		{"iam", iam.MigrationDocuments},
		{"system resource layout", systemresourcelayout.MigrationDocuments},
		{"billing", billing.MigrationDocuments},
		{"usage", usage.MigrationDocuments},
		{"platform async", platformasync.MigrationDocuments},
		{"platform backup receipt", backupreceipt.MigrationDocuments},
		{"analytics", analytics.MigrationDocuments},
		{"catalog", catalog.MigrationDocuments},
		{"work", work.MigrationDocuments},
		{"documents", documents.MigrationDocuments},
		{"notifications", notifications.MigrationDocuments},
		{"support", support.MigrationDocuments},
		{"audit", audit.MigrationDocuments},
	}
}
