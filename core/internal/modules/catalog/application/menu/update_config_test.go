package menu

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestUpdateConfigRejectsMissingID(t *testing.T) {
	_, err := (*Application)(nil).UpdateConfig(context.Background(), platformtx.TenantScope{},
		ConfigWriteRequest{Name: "Menu", Config: json.RawMessage(`{"groups":[],"ungrouped":[]}`),
			Active: true, IdempotencyKey: "catalog-menu-config-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("update menu config error = %v", err)
	}
}

func TestUpdateConfigRejectsNonV7ID(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	_, err := (*Application)(nil).UpdateConfig(context.Background(), platformtx.TenantScope{},
		ConfigWriteRequest{ID: id, Name: "Menu",
			Config: json.RawMessage(`{"groups":[],"ungrouped":[]}`),
			Active: true, IdempotencyKey: "catalog-menu-config-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("update menu config error = %v", err)
	}
}
