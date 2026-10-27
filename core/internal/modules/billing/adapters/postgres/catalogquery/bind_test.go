package catalogquery

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestBindRequiresLiveGlobalSnapshotScope(t *testing.T) {
	if _, err := Bind(platformtx.GlobalScope{}); !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("invalid global scope error = %v", err)
	}
}

func TestQueriesRequireBoundSnapshotContext(t *testing.T) {
	storage := &store{query: func(context.Context, string, ...any) (rows, error) {
		t.Fatal("query must not run outside bound scope")
		return nil, nil
	}, allowed: func(context.Context) bool { return false }}
	ctx := context.Background()
	if _, err := storage.ListOfferings(ctx, catalogapp.ListStoreQuery{}); !errors.Is(
		err, catalogapp.ErrConfiguration,
	) {
		t.Fatalf("list error = %v", err)
	}
	if _, _, err := storage.GetOffering(ctx, catalogapp.GetStoreQuery{}); !errors.Is(
		err, catalogapp.ErrConfiguration,
	) {
		t.Fatalf("get error = %v", err)
	}
	if _, _, err := storage.ReadActiveCampaign(
		ctx, catalogapp.CampaignStoreQuery{},
	); !errors.Is(err, catalogapp.ErrConfiguration) {
		t.Fatalf("campaign error = %v", err)
	}
}
