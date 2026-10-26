package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsUnownedScope(t *testing.T) {
	value, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = value.RetireAuthenticator(
		context.Background(), platformtx.TenantScope{}, lifecycle.RevisionCommand{},
	)
	if !errors.Is(err, lifecycle.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
