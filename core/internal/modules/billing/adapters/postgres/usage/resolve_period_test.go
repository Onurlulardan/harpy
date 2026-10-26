package usage

import (
	"context"
	"errors"
	"testing"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestResolvePeriodRejectsUnboundStore(t *testing.T) {
	key, err := entitlement.NewKey("documents.storage_bytes")
	if err != nil {
		t.Fatal(err)
	}
	var storage *store
	if _, err = storage.ResolvePeriod(context.Background(), key); !errors.Is(
		err, usageapp.ErrConfiguration,
	) {
		t.Fatalf("expected configuration error, got %v", err)
	}
}
