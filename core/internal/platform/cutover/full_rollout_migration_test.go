package cutover

import (
	"path"
	"strings"
	"testing"
)

func fullRolloutMigrationSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026071823000") {
			name := path.Base(document.Path())
			sources[name[:14]] = string(document.Content())
		}
	}
	if len(sources) != 4 {
		t.Fatalf("full rollout migration count = %d, want 4", len(sources))
	}
	return sources
}
