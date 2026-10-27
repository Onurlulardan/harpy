package billing

import (
	"strings"
	"testing"
)

func TestPromotionCompatibilitySelectsNewestCreatedCampaign(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, document := range documents {
		if !strings.Contains(document.Path(),
			"20260717010306_guard_billing_promotion_v1") {
			continue
		}
		found++
		content := string(document.Content())
		if !strings.Contains(content,
			"ORDER BY root.created_at DESC, root.id ASC") {
			t.Fatalf("%s does not preserve V1 created_at ordering", document.Path())
		}
		if strings.Contains(content,
			"ORDER BY version_row.priority DESC, version_row.id ASC") {
			t.Fatalf("%s still selects campaigns by priority", document.Path())
		}
	}
	if found != 2 {
		t.Fatalf("promotion guard migration documents=%d want=2", found)
	}
}
