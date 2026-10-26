package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func TestMaterializeLegacyIdentityRequiresBoundSnapshotScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, _, err := storage.MaterializeLegacyIdentity(
		context.Background(), catalogapp.LegacyIdentityStoreQuery{},
	)
	if !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("materialize error=%v", err)
	}
}
