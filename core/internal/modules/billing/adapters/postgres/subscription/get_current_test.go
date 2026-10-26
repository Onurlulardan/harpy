package subscription

import (
	"context"
	"errors"
	"testing"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func TestGetCurrentRequiresBoundSnapshotScope(t *testing.T) {
	storage := &queryStore{
		queryRow: func(context.Context, string, ...any) row {
			t.Fatal("query must not run outside snapshot scope")
			return nil
		},
		allowed:  func(context.Context) bool { return false },
		notFound: func(error) bool { return false },
	}
	_, err := storage.GetCurrent(context.Background())
	if !errors.Is(err, subscriptionapp.ErrQueryConfiguration) {
		t.Fatalf("get current error = %v", err)
	}
}
