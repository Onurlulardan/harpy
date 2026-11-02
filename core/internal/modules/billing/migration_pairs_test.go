package billing

import (
	"strings"
	"testing"
)

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
