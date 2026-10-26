package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateRejectsUnownedScope(t *testing.T) {
	value, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = value.CreatePortalSetup(
		context.Background(), platformtx.TenantScope{}, lifecycle.PortalSetupStartCommand{},
	)
	if !errors.Is(err, lifecycle.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
