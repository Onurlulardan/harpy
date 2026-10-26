package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func TestGetOfferingRequiresBoundSnapshotScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, _, err := storage.GetOffering(context.Background(), catalogapp.GetStoreQuery{})
	if !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("get error = %v", err)
	}
}
