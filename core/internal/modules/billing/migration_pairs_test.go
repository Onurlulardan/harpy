package billing

import (
	"bytes"
	"path"
	"strings"
	"testing"
)

func TestPlatformMigrationsAreForwardOnlyExactOriginPairs(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pairs := make(map[string]map[string][]byte)
	for _, document := range documents {
		origin, key := billingMigrationOrigin(document.Path())
		if origin == "" || key == "" {
			t.Fatalf("unrecognized billing migration path %q", document.Path())
		}
		if pairs[key] == nil {
			pairs[key] = make(map[string][]byte)
		}
		if _, exists := pairs[key][origin]; exists {
			t.Fatalf("duplicate billing migration %s/%s", origin, key)
		}
		content := document.Content()
		pairs[key][origin] = content
		assertBillingMigrationShape(t, document.Path(), content)
	}
	if len(pairs) != 208 {
		t.Fatalf("billing migration pair count = %d, want 208", len(pairs))
	}
	for key, pair := range pairs {
		fresh, hasFresh := pair["fresh"]
		adopted, hasAdopted := pair["v1_adopted"]
		if !hasFresh || !hasAdopted || !bytes.Equal(fresh, adopted) {
			t.Errorf("billing migration %s is not an exact origin pair", key)
		}
	}
}

func billingMigrationOrigin(documentPath string) (string, string) {
	name := path.Base(documentPath)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		suffix := "." + origin + ".sql"
		if strings.HasSuffix(name, suffix) &&
			strings.Contains(documentPath, "/"+origin+"/") {
			return origin, strings.TrimSuffix(name, suffix)
		}
	}
	return "", ""
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
