package cutover

import (
	"path"
	"strings"
	"testing"
)

func systemResourceLayoutMigrationSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/fresh/202607160058") {
			continue
		}
		name := path.Base(document.Path())
		sources[name[:14]] = string(document.Content())
	}
	if len(sources) != 7 {
		t.Fatalf("forward migration count = %d, want 7", len(sources))
	}
	return sources
}
