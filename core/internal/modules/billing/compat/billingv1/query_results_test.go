package billingv1

import (
	"encoding/json"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestCatalogRootIDsDeduplicatesVersionsAcrossCollectedPages(t *testing.T) {
	items := []json.RawMessage{
		json.RawMessage(`{"discount_id":"older-root","version":2}`),
		json.RawMessage(`{"discount_id":"newer-root","version":1}`),
		json.RawMessage(`{"discount_id":"older-root","version":1}`),
	}
	identities, err := catalogRootIDs(items, "discount_id")
	if err != nil || len(identities) != 2 ||
		identities[0] != "older-root" || identities[1] != "newer-root" {
		t.Fatalf("identities=%v error=%v", identities, err)
	}
	if _, err = catalogRootIDs(
		[]json.RawMessage{json.RawMessage(`{"version":1}`)}, "discount_id",
	); !errors.Is(err, ErrPaginationEvidence) {
		t.Fatalf("missing root identity error=%v", err)
	}
}

func TestPromotionListsSortByCreatedAtDescending(t *testing.T) {
	discounts, err := sortDiscountViews([]billingapi.DiscountView{
		{ID: "older", CreatedAt: "2026-07-01T00:00:00.000000Z"},
		{ID: "newer", CreatedAt: "2026-07-02T00:00:00.000000Z"},
	})
	if err != nil || discounts[0].ID != "newer" || discounts[1].ID != "older" {
		t.Fatalf("discounts=%v error=%v", discounts, err)
	}
	campaigns, err := sortCampaignViews([]billingapi.CampaignView{
		{ID: "older", CreatedAt: "2026-07-01T00:00:00.000000Z"},
		{ID: "newer", CreatedAt: "2026-07-02T00:00:00.000000Z"},
	})
	if err != nil || campaigns[0].ID != "newer" || campaigns[1].ID != "older" {
		t.Fatalf("campaigns=%v error=%v", campaigns, err)
	}
	if _, err = sortCampaignViews([]billingapi.CampaignView{{CreatedAt: "bad"}}); !errors.Is(err, ErrSnapshotRequired) {
		t.Fatalf("invalid campaign instant error=%v", err)
	}
}
