package offering

import (
	"context"
	"errors"
	"testing"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

func TestPublishRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Publish(
		context.Background(), offeringapp.CreateVersionMutation{},
	)
	if !errors.Is(err, offeringapp.ErrConfiguration) {
		t.Fatalf("publish error = %v", err)
	}
}
