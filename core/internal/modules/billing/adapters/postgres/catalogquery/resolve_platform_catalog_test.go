package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func TestResolvePlatformCatalogRequiresBoundSnapshotScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, _, err := storage.ResolvePlatformCatalog(
		context.Background(), catalogapp.PlatformResolveStoreQuery{},
	)
	if !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("resolve error = %v", err)
	}
}
