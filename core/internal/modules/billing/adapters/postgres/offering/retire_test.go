package offering

import (
	"context"
	"errors"
	"testing"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

func TestRetireRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Retire(context.Background(), offeringapp.RetireMutation{})
	if !errors.Is(err, offeringapp.ErrConfiguration) {
		t.Fatalf("retire error = %v", err)
	}
}
