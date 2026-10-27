package billing

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformMigrationsAreForwardOnlyExactOriginPairs(t *testing.T) {
	freshPaths, err := filepath.Glob(filepath.Join(
		"migrations", "platform", "fresh", "*.fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	if len(freshPaths) != 208 {
		t.Fatalf("billing migration pair count = %d, want 208", len(freshPaths))
	}
	seen := make(map[string]struct{}, len(freshPaths))
	for _, freshPath := range freshPaths {
		key := strings.TrimSuffix(filepath.Base(freshPath), ".fresh.sql")
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate billing migration fresh/%s", key)
		}
		seen[key] = struct{}{}
		fresh, readErr := os.ReadFile(freshPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		adoptedPath := filepath.Join(
			"migrations", "platform", "v1_adopted", key+".v1_adopted.sql",
		)
		adopted, readErr := os.ReadFile(adoptedPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !bytes.Equal(fresh, adopted) {
			t.Errorf("billing migration %s is not an exact origin pair", key)
		}
		assertBillingMigrationShape(t, freshPath, fresh)
		assertBillingMigrationShape(t, adoptedPath, adopted)
	}
}

func assertBillingMigrationShape(t *testing.T, name string, content []byte) {
	t.Helper()
	text := string(content)
	for _, forbidden := range []string{
		"-- +goose Down", "CREATE EXTENSION", "DO $$", "DO $block$",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("billing migration %s contains %q", name, forbidden)
		}
	}
	beginCount := strings.Count(text, "-- +goose StatementBegin")
	endCount := strings.Count(text, "-- +goose StatementEnd")
	if beginCount != endCount || beginCount > 1 {
		t.Errorf("billing migration %s has invalid statement block count %d/%d",
			name, beginCount, endCount)
	}
}
