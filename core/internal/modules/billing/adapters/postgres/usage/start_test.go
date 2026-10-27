package usage

import (
	"context"
	"errors"
	"testing"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func TestStartRejectsUnboundStore(t *testing.T) {
	var storage *store
	_, err := storage.Start(context.Background(), usageapp.StartExecution{}, nil)
	if !errors.Is(err, usageapp.ErrConfiguration) {
		t.Fatalf("expected configuration error, got %v", err)
	}
}
