//go:build integration

package integrationfixture

import "github.com/Onurlulardan/harpy/core/internal/platform/migration"

// LoadAsyncMigrationDocuments returns no per-module SQL. Tenant and platform
// schema now come from the Goose baseline.
func LoadAsyncMigrationDocuments() ([]migration.Document, error) {
	return nil, nil
}
