package dashboard

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestGetRejectsUnboundScope(t *testing.T) {
	_, err := (&Application{}).Get(context.Background(), platformtx.TenantScope{}, dashboardFixtureID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("get error = %v", err)
	}
}
