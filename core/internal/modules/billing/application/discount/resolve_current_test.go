package discount

import (
	"context"
	"errors"
	"testing"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolveCurrentRejectsUnboundScopeBeforeStorage(t *testing.T) {
	store := &rejectingEligibilityStore{}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	queries, err := NewQueries(store, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	code, _ := discountdomain.NormalizeCode("launch")
	_, err = queries.ResolveCurrent(
		context.Background(), platformtx.TenantScope{}, string(code),
	)
	if !errors.Is(err, ErrQueryScope) || store.called {
		t.Fatalf("resolve error=%v called=%t", err, store.called)
	}
}
