package usage

import (
	"context"
	"errors"
	"testing"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func TestExpireRejectsUnboundStore(t *testing.T) {
	var storage *store
	_, err := storage.Expire(context.Background(), usageapp.ExpireExecution{}, nil)
	if !errors.Is(err, usageapp.ErrConfiguration) {
		t.Fatalf("expected configuration error, got %v", err)
	}
}
