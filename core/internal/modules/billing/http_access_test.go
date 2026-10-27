package billing

import (
	"context"
	"errors"
	"net/http"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type accessVerifierFunc func(context.Context, AccessRequest) (billingapi.Access, error)

func (verify accessVerifierFunc) Verify(
	ctx context.Context,
	request AccessRequest,
) (billingapi.Access, error) {
	return verify(ctx, request)
}

func TestAccessForPassesRawAssertionsAndUsesOnlyVerifiedAccess(t *testing.T) {
	var captured AccessRequest
	handler := &Handler{access: accessVerifierFunc(func(
		_ context.Context,
		request AccessRequest,
	) (billingapi.Access, error) {
		captured = request
		return billingapi.NewTenantAccess(
			uuid.MustParse("00000000-0000-4000-8000-000000000001"),
			uuid.MustParse("00000000-0000-4000-8000-000000000002"),
			5,
			request.Permission,
		)
	})}
	request, err := http.NewRequest(http.MethodGet, "/api/billing/v2/usage", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header["Authorization"] = []string{"Bearer first", "Bearer second"}
	request.Header["X-Tenant-Id"] = []string{"tenant-a", "tenant-b"}
	request.Header["X-Tenant-Slug"] = []string{"forge"}
	access, err := handler.accessFor(request, billingapi.PermissionBillingUsageRead)
	if err != nil || !access.ValidFor(billingapi.PermissionBillingUsageRead) {
		t.Fatalf("verified access=%#v err=%v", access, err)
	}
	if len(captured.AuthorizationValues) != 2 || len(captured.TenantIDValues) != 2 ||
		len(captured.TenantSlugValues) != 1 {
		t.Fatalf("raw assertions changed: %#v", captured)
	}
}

func TestAccessForRejectsVerifierResultForAnotherPermission(t *testing.T) {
	handler := &Handler{access: accessVerifierFunc(func(
		context.Context,
		AccessRequest,
	) (billingapi.Access, error) {
		return billingapi.NewTenantAccess(
			uuid.MustParse("00000000-0000-4000-8000-000000000001"),
			uuid.MustParse("00000000-0000-4000-8000-000000000002"),
			5,
			billingapi.PermissionBillingRead,
		)
	})}
	request, err := http.NewRequest(http.MethodPost, "/resource", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.accessFor(
		request, billingapi.PermissionBillingManage,
	); !errors.Is(err, billingapi.ErrForbidden) {
		t.Fatalf("wrong permission result returned %v", err)
	}
}
