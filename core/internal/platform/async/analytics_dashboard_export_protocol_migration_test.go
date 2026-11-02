package async

import (
	"path"
	"strconv"
	"strings"
	"testing"
)

const analyticsDashboardExportMigrationStart = int64(20260802020000)
const analyticsDashboardExportMigrationEnd = int64(20260802020049)

func analyticsDashboardExportFreshMigrations(t *testing.T) map[string][]byte {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	fresh := map[string][]byte{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasSuffix(name, ".fresh.sql") {
			continue
		}
		version, parseErr := strconv.ParseInt(name[:14], 10, 64)
		if parseErr != nil || version < analyticsDashboardExportMigrationStart ||
			version > analyticsDashboardExportMigrationEnd {
			continue
		}
		key := strings.TrimSuffix(name, ".fresh.sql")
		if _, exists := fresh[key]; exists &&
			strings.Contains(document.Path(), "/supersessions/") {
			continue
		}
		fresh[key] = document.Content()
	}
	return fresh
}
