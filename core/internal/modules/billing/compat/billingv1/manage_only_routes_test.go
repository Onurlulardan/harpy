package billingv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type manageOnlyAccessStub struct {
	manageCalls int
	readCalls   int
}

func (stub *manageOnlyAccessStub) Verify(
	_ context.Context, request billingapi.AccessRequest,
) (billingapi.Access, error) {
	if request.Permission != billingapi.PermissionPlatformBillingManage {
		stub.readCalls++
		return billingapi.Access{}, billingapi.ErrForbidden
	}
	stub.manageCalls++
	return billingapi.NewGlobalAccess(uuid.New(), 1, request.Permission)
}

type manageOnlyReportsStub struct{ billingapi.PlatformReports }

func (manageOnlyReportsStub) ListSubscriptions(
	context.Context, billingapi.Query,
) (billingapi.Page[billingapi.SubscriptionView], error) {
	return billingapi.Page[billingapi.SubscriptionView]{}, billingapi.ErrNotFound
}

type manageOnlyRouteCase struct {
	name, method, body string
	want               int
	serve              func(*Handler, http.ResponseWriter, *http.Request)
}

func TestMutationRoutesDoNotRequireSecondReadPermission(t *testing.T) {
	t.Parallel()
	routes := []manageOnlyRouteCase{
		{"plan-update", http.MethodPut, `{}`, http.StatusNotFound,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.updatePlan(w, r) }},
		{"plan-delete", http.MethodDelete, ``, http.StatusOK,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.deletePlan(w, r) }},
		{"coupon-create", http.MethodPost,
			`{"code":"SAVE10","name":"Save","discount_type":"percentage","discount_value":10}`,
			http.StatusNotFound,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.createCoupon(w, r) }},
		{"coupon-update", http.MethodPut, `{}`, http.StatusNotFound,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.updateCoupon(w, r) }},
		{"coupon-delete", http.MethodDelete, ``, http.StatusOK,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.deleteCoupon(w, r) }},
		{"campaign-update", http.MethodPut, `{}`, http.StatusNotFound,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.updateCampaign(w, r) }},
		{"campaign-delete", http.MethodDelete, ``, http.StatusOK,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.deleteCampaign(w, r) }},
		{"subscription-update", http.MethodPut, `{"status":"active"}`, http.StatusNotFound,
			func(h *Handler, w http.ResponseWriter, r *http.Request) { h.updateSubscription(w, r) }},
	}
	for _, route := range routes {
		route := route
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			access := &manageOnlyAccessStub{}
			handler := manageOnlyHandler(access)
			request := httptest.NewRequest(
				route.method, "/", strings.NewReader(route.body),
			)
			request.SetPathValue("id", uuid.NewString())
			request.Header.Set("Idempotency-Key", "manage-only-test")
			response := httptest.NewRecorder()
			route.serve(handler, response, request)
			if response.Code != route.want {
				t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
			}
			if access.manageCalls != 1 || access.readCalls != 0 {
				t.Fatalf("manage/read calls = %d/%d", access.manageCalls, access.readCalls)
			}
		})
	}
}

func manageOnlyHandler(access AccessVerifier) *Handler {
	return &Handler{
		access:   access,
		identity: deleteIdentityStub{},
		catalog:  deleteCatalogStub{discountErr: billingapi.ErrNotFound},
		reports:  manageOnlyReportsStub{},
		clock:    time.Now,
		query:    billingapi.NewQueryDecoder(),
	}
}
