package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func TestReadActiveCampaignRequiresBoundSnapshotScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, _, err := storage.ReadActiveCampaign(
		context.Background(), catalogapp.CampaignStoreQuery{},
	)
	if !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("campaign error = %v", err)
	}
}
