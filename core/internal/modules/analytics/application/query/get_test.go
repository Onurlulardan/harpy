package query

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestGetRejectsUnboundTenantScope(t *testing.T) {
	_, err := (&Application{}).Get(context.Background(), platformtx.TenantScope{}, queryFixtureID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("get error = %v", err)
	}
}
