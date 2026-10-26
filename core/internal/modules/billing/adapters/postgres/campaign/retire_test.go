package campaign

import (
	"context"
	"errors"
	"testing"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
)

func TestRetireRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Retire(context.Background(), campaignapp.RetireMutation{})
	if !errors.Is(err, campaignapp.ErrConfiguration) {
		t.Fatalf("retire error = %v", err)
	}
}
