package billing

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
	operations ...string,
) (billingapi.Access, error) {
	if r == nil {
		return billingapi.Access{}, billingapi.ErrAuthentication
	}
	operation := ""
	if len(operations) == 1 {
		operation = operations[0]
	}
	access, err := handler.access.Verify(r.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		Permission:          permission,
		Operation:           operation,
	})
	if err != nil || !access.ValidFor(permission) {
		if err != nil {
			return billingapi.Access{}, err
		}
		return billingapi.Access{}, billingapi.ErrForbidden
	}
	return access, nil
}
