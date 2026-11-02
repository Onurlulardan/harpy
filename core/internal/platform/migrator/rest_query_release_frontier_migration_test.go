package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

const iamRestQueryReleaseFrontierMigration = "20260730113100_assert_iam_rest_query_search_release_frontier_v2"

func readIAMNamedPlatformMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	if origin != "fresh" {
		t.Fatalf("fresh migration %s did not load", name)
	}
	path := filepath.Join("migrations", "fresh", name+".fresh.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
