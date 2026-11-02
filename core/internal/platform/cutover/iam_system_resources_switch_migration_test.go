package cutover

import (
	"path"
	"strings"
	"testing"
)

func iamSystemResourcesSwitchSources(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/2026071602000") {
			name := path.Base(document.Path())
			sources[name[:14]] = string(document.Content())
		}
	}
	if len(sources) != 2 {
		t.Fatalf("finalized switch migration count = %d, want 2", len(sources))
	}
	return sources
}
