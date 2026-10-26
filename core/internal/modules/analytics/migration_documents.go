package analytics

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func MigrationDocuments() ([]migration.Document, error) {
	return nil, fmt.Errorf("harpy schema migrations are not embedded")
}
