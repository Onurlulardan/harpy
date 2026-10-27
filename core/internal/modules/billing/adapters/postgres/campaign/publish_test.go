package campaign

import (
	"context"
	"errors"
	"testing"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
)

func TestPublishRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Publish(context.Background(), campaignapp.CreateVersionMutation{})
	if !errors.Is(err, campaignapp.ErrConfiguration) {
		t.Fatalf("publish error = %v", err)
	}
}
