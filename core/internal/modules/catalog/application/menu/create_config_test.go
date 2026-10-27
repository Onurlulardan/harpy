package menu

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateConfigRejectsInvalidName(t *testing.T) {
	_, err := (*Application)(nil).CreateConfig(context.Background(), platformtx.TenantScope{},
		ConfigWriteRequest{Name: " ", Config: json.RawMessage(`{"groups":[],"ungrouped":[]}`),
			IdempotencyKey: "catalog-menu-config-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("create menu config error = %v", err)
	}
}

func TestValidConfigJSONAcceptsDesignerPayload(t *testing.T) {
	raw := json.RawMessage(`{"groups":[{"id":"g1","title":"HR","items":[{"form_slug":"people"}],
		"groups":[]}],"ungrouped":[{"form_slug":"orders"}]}`)
	if !validConfigJSON(raw) || countConfigItems(raw) != 2 {
		t.Fatal("designer menu JSON must accept nested groups and form_slug items")
	}
}

func TestValidConfigJSONRejectsNonArrayGroups(t *testing.T) {
	if validConfigJSON(json.RawMessage(`{"groups":{},"ungrouped":[]}`)) {
		t.Fatal("menu groups must be an array")
	}
}
