package billing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformMigrationsAreForwardOnlyFresh(t *testing.T) {
	freshPaths, err := filepath.Glob(filepath.Join(
		"migrations", "platform", "fresh", "*.fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	if len(freshPaths) != 208 {
		t.Fatalf("billing fresh migration count = %d, want 208", len(freshPaths))
	}
	seen := make(map[string]struct{}, len(freshPaths))
	for _, freshPath := range freshPaths {
		key := strings.TrimSuffix(filepath.Base(freshPath), ".fresh.sql")
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate billing migration fresh/%s", key)
		}
		seen[key] = struct{}{}
		fresh, readErr := os.ReadFile(freshPath)
		if readErr != nil || len(fresh) == 0 {
			t.Fatalf("billing fresh migration %s did not load: %v", key, readErr)
		}
		assertBillingMigrationShape(t, freshPath, fresh)
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
