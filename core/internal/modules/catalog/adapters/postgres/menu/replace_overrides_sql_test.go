package menu

import (
	"strings"
	"testing"
)

func TestReplaceOverridesDoNotRequireDeleteGrant(t *testing.T) {
	if !strings.Contains(revealItemOverrideSQL, "UPDATE catalog_menu_overrides") ||
		strings.Contains(revealItemOverrideSQL, "DELETE") {
		t.Fatal("visible menu items must UPDATE overrides without DELETE")
	}
	if !strings.Contains(insertHiddenOverrideSQL, "ON CONFLICT (item_id)") {
		t.Fatal("hidden menu override must upsert")
	}
}
