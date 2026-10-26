package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func TestListOfferingsRequiresBoundSnapshotScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.ListOfferings(context.Background(), catalogapp.ListStoreQuery{})
	if !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("list error = %v", err)
	}
}
