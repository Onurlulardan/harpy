package usage

import (
	"context"
	"errors"
	"testing"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func TestGetRequiresBoundSnapshotScope(t *testing.T) {
	storage := &queryStore{
		queryRow: func(context.Context, string, ...any) row {
			t.Fatal("query must not run outside snapshot scope")
			return nil
		},
		query: func(context.Context, string, ...any) (queryRows, error) {
			t.Fatal("query must not run outside snapshot scope")
			return nil, nil
		},
		allowed:  func(context.Context) bool { return false },
		notFound: func(error) bool { return false },
	}
	_, err := storage.Get(context.Background())
	if !errors.Is(err, usageapp.ErrQueryConfiguration) {
		t.Fatalf("get usage error = %v", err)
	}
}
