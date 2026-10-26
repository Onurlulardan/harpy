package billingv1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type capturingAccess struct {
	request AccessRequest
	calls   int
}

func (access *capturingAccess) Verify(
	_ context.Context,
	request AccessRequest,
) (billingapi.Access, error) {
	access.calls++
	access.request = request
	return billingapi.Access{}, billingapi.ErrForbidden
}

func TestAuthorizeForwardsExactHeadersPermissionAndOperation(t *testing.T) {
	t.Parallel()
	verifier := &capturingAccess{}
	handler := &Handler{access: verifier}
	request := httptest.NewRequest("GET", "/api/billing/subscription", nil)
	request.Header.Add("Authorization", "Bearer first")
	request.Header.Add("Authorization", "Bearer second")
	request.Header.Add("X-Tenant-ID", "tenant-id")
	request.Header.Add("X-Tenant-Slug", "tenant-slug")
	_, err := handler.authorize(
		request, "billing.route.subscription-read", billingapi.PermissionBillingRead,
	)
	if !errors.Is(err, billingapi.ErrForbidden) || verifier.calls != 1 {
		t.Fatalf("authorization calls=%d error=%v", verifier.calls, err)
	}
	got := verifier.request
	if got.Permission != billingapi.PermissionBillingRead ||
		got.Operation != "billing.subscription.read" ||
		len(got.AuthorizationValues) != 2 ||
		len(got.TenantIDValues) != 1 || len(got.TenantSlugValues) != 1 {
		t.Fatalf("authorization request = %#v", got)
	}
}

func TestAuthorizeRejectsUnknownRouteBeforeAccessPort(t *testing.T) {
	t.Parallel()
	verifier := &capturingAccess{}
	handler := &Handler{access: verifier}
	_, err := handler.authorize(
		httptest.NewRequest("GET", "/", nil),
		"billing.route.unknown", billingapi.PermissionBillingRead,
	)
	if !errors.Is(err, billingapi.ErrForbidden) || verifier.calls != 0 {
		t.Fatalf("unknown route calls=%d error=%v", verifier.calls, err)
	}
}
