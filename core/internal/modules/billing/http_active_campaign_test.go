package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestActiveCampaignPreservesAuthoritativeDeterministicSnapshot(t *testing.T) {
	view := &billingapi.CampaignView{
		ID:  "00000000-0000-4000-8000-000000000020",
		Key: "summer", Revision: 4,
		VersionID: "00000000-0000-4000-8000-000000000021", Version: 3,
		State: "published", Name: "Summer", Priority: 100,
		StartsAt: "2026-07-01T00:00:00.000000Z",
		Discounts: []billingapi.DiscountView{
			{ID: "00000000-0000-4000-8000-000000000030", Code: "FIRST", Priority: 20},
			{ID: "00000000-0000-4000-8000-000000000031", Code: "SECOND", Priority: 10},
		},
	}
	handler := newHandlerFixture(t, publicStub{active: func(
		context.Context,
	) (*billingapi.CampaignView, error) {
		return view, nil
	}}, nil)
	response := httptest.NewRecorder()
	handler.getActiveCampaign(
		response,
		httptest.NewRequest(http.MethodGet, "/api/billing/v2/promotions/active", nil),
	)
	var envelope struct {
		Data billingapi.CampaignView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || envelope.Data.VersionID != view.VersionID ||
		len(envelope.Data.Discounts) != 2 || envelope.Data.Discounts[0].Code != "FIRST" ||
		envelope.Data.Discounts[1].Code != "SECOND" {
		t.Fatalf("authoritative campaign snapshot changed: %#v", envelope.Data)
	}
}

func TestActiveCampaignReturnsJSONNullWhenSelectionIsEmpty(t *testing.T) {
	handler := newHandlerFixture(t, publicStub{active: func(
		context.Context,
	) (*billingapi.CampaignView, error) {
		return nil, nil
	}}, nil)
	response := httptest.NewRecorder()
	handler.getActiveCampaign(
		response,
		httptest.NewRequest(http.MethodGet, "/api/billing/v2/promotions/active", nil),
	)
	var envelope struct {
		Data *billingapi.CampaignView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || envelope.Data != nil {
		t.Fatalf("empty campaign response mismatch: %d %#v", response.Code, envelope.Data)
	}
}
