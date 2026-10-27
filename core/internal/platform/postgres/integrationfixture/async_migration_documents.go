//go:build integration

package integrationfixture

import (
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

// LoadAsyncMigrationDocuments returns the runtime-selected async migration set.
func LoadAsyncMigrationDocuments() ([]migration.Document, error) {
	return platformasync.MigrationDocuments()
}
