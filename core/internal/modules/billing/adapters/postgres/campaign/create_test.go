package campaign

import (
	"context"
	"errors"
	"testing"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
)

func TestCreateRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Create(context.Background(), campaignapp.CreateMutation{})
	if !errors.Is(err, campaignapp.ErrConfiguration) {
		t.Fatalf("create error = %v", err)
	}
}
