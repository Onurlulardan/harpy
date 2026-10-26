package billingv1

import (
	"encoding/json"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestLegacyCouponWireGoldenJSON(t *testing.T) {
	basisPoints := int64(1500)
	wire, err := legacyCoupon(billingapi.DiscountView{
		ID:        "019c92a0-0000-7000-8000-000000000074",
		LegacyID:  "74000000-0000-4000-8000-000000000074",
		CreatedBy: "75000000-0000-4000-8000-000000000075",
		Code:      "SAVE15", Name: "Save fifteen", BasisPoints: &basisPoints,
		Duration: "once", UsedCount: 7, State: "active", Available: true,
		ValidStart: "2026-07-01T00:00:00.000000Z",
		CreatedAt:  "2026-06-01T00:00:00.000000Z",
		UpdatedAt:  "2026-06-02T03:04:05.123456Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"74000000-0000-4000-8000-000000000074","code":"SAVE15","name":"Save fifteen","discount_type":"percentage","discount_value":15,"max_uses":null,"used_count":7,"max_uses_per_tenant":1,"valid_from":"2026-07-01T00:00:00Z","applicable_plan_ids":[],"duration":"once","is_active":true,"created_by":"75000000-0000-4000-8000-000000000075","created_at":"2026-06-01T00:00:00Z","updated_at":"2026-06-02T03:04:05.123456Z"}`
	if string(document) != want {
		t.Fatalf("coupon JSON=%s\nwant=%s", document, want)
	}
}

func TestLegacyCampaignWireGoldenJSON(t *testing.T) {
	wire, err := legacyCampaign(billingapi.CampaignView{
		ID:        "019c92a0-0000-7000-8000-000000000084",
		LegacyID:  "84000000-0000-4000-8000-000000000084",
		CreatedBy: "85000000-0000-4000-8000-000000000085",
		Key:       "summer", Name: "Summer", AutoApply: true,
		State: "active", Available: true,
		StartsAt:  "2026-07-01T00:00:00.000000Z",
		CreatedAt: "2026-06-01T00:00:00.000000Z",
		UpdatedAt: "2026-06-02T03:04:05.123456Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"84000000-0000-4000-8000-000000000084","name":"Summer","slug":"summer","start_date":"2026-07-01T00:00:00Z","auto_apply":true,"is_active":true,"created_by":"85000000-0000-4000-8000-000000000085","created_at":"2026-06-01T00:00:00Z","updated_at":"2026-06-02T03:04:05.123456Z"}`
	if string(document) != want {
		t.Fatalf("campaign JSON=%s\nwant=%s", document, want)
	}
}
