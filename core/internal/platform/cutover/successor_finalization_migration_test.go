package cutover

import (
	"path"
	"strings"
	"testing"
)

func successorFinalizationMigrationPair(
	t *testing.T,
	version string,
) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pair := map[string]string{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if strings.HasPrefix(name, version+"_") && strings.HasSuffix(name, ".fresh.sql") {
			pair["fresh"] = string(document.Content())
		}
	}
	if pair["fresh"] == "" {
		t.Fatalf("fresh migration %s did not load", version)
	}
	return pair
}
