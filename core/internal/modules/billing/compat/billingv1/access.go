package billingv1

import (
	"context"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type AccessRequest = billingapi.AccessRequest

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (billingapi.Access, error)
}

func (handler *Handler) accessFor(
	r *http.Request,
	permission billingapi.Permission,
	operation string,
) (billingapi.Access, error) {
	if r == nil {
		return billingapi.Access{}, billingapi.ErrAuthentication
	}
	access, err := handler.access.Verify(r.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		Permission:          permission,
		Operation:           operation,
	})
	if err != nil {
		return billingapi.Access{}, err
	}
	if !access.ValidFor(permission) {
		return billingapi.Access{}, billingapi.ErrForbidden
	}
	return access, nil
}
