package billingv1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type deleteRoute func(*Handler, http.ResponseWriter, *http.Request)

type deleteRouteCase struct {
	name    string
	kind    billingapi.LegacyIdentityKind
	message string
	route   deleteRoute
}

func TestPromotionDeletesAreIdempotentAndKindFenced(t *testing.T) {
	t.Parallel()
	tests := []deleteRouteCase{
		{"coupon", billingapi.LegacyIdentityDiscount, "Coupon deleted", callDeleteCoupon},
		{"campaign", billingapi.LegacyIdentityCampaign, "Campaign deleted", callDeleteCampaign},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			identifier := uuid.NewString()
			assertDeleteStatus(t, test, identifier, deleteIdentityStub{},
				deleteCatalogStub{}, http.StatusOK)
			wrongKind := billingapi.LegacyIdentityOffering
			assertDeleteStatus(t, test, identifier, deleteIdentityStub{
				identifier: identifier, kind: wrongKind, canonical: uuid.NewString(),
			}, deleteCatalogStub{}, http.StatusBadRequest)
			assertDeleteStatus(t, test, "not-a-uuid", deleteIdentityStub{},
				deleteCatalogStub{}, http.StatusBadRequest)
		})
	}
}

func TestPromotionDeletesAcceptRetiredOrMissingAdoptedResource(t *testing.T) {
	t.Parallel()
	identifier, canonical := uuid.NewString(), uuid.NewString()
	tests := []struct {
		deleteRouteCase
		catalog deleteCatalogStub
	}{
		{
			deleteRouteCase{"coupon-retired", billingapi.LegacyIdentityDiscount,
				"Coupon deleted", callDeleteCoupon},
			deleteCatalogStub{discount: billingapi.DiscountView{State: "retired"}},
		},
		{
			deleteRouteCase{"coupon-deleted", billingapi.LegacyIdentityDiscount,
				"Coupon deleted", callDeleteCoupon},
			deleteCatalogStub{discountErr: billingapi.ErrNotFound},
		},
		{
			deleteRouteCase{"campaign-retired", billingapi.LegacyIdentityCampaign,
				"Campaign deleted", callDeleteCampaign},
			deleteCatalogStub{campaign: billingapi.CampaignView{State: "retired"}},
		},
		{
			deleteRouteCase{"campaign-deleted", billingapi.LegacyIdentityCampaign,
				"Campaign deleted", callDeleteCampaign},
			deleteCatalogStub{campaignErr: billingapi.ErrNotFound},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assertDeleteStatus(t, test.deleteRouteCase, identifier, deleteIdentityStub{
				identifier: identifier, kind: test.kind, canonical: canonical,
			}, test.catalog, http.StatusOK)
		})
	}
}

func assertDeleteStatus(
	t *testing.T,
	test deleteRouteCase,
	identifier string,
	identity deleteIdentityStub,
	catalog deleteCatalogStub,
	want int,
) {
	t.Helper()
	handler := &Handler{access: deleteAccessStub{}, identity: identity, catalog: catalog}
	request := httptest.NewRequest(http.MethodDelete, "/", nil)
	request.SetPathValue("id", identifier)
	response := httptest.NewRecorder()
	test.route(handler, response, request)
	if response.Code != want {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if want == http.StatusOK {
		var body map[string]string
		_ = json.Unmarshal(response.Body.Bytes(), &body)
		if body["message"] != test.message {
			t.Fatalf("message = %q", body["message"])
		}
	}
}

func callDeleteCoupon(handler *Handler, w http.ResponseWriter, r *http.Request) {
	handler.deleteCoupon(w, r)
}

func callDeleteCampaign(handler *Handler, w http.ResponseWriter, r *http.Request) {
	handler.deleteCampaign(w, r)
}
