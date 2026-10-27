package audithttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

func (handler *Handler) accessFor(request *http.Request, operation string) (Access, error) {
	if !handler.Ready() || request == nil {
		return Access{}, application.ErrUnavailable
	}
	access, err := handler.dependencies.Access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), request.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), request.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), request.Header.Values("X-Tenant-Slug")...),
		Permission:          application.PermissionRead, Operation: operation,
	})
	if err != nil {
		return Access{}, err
	}
	if !access.ValidFor(application.PermissionRead, operation) {
		return Access{}, application.ErrForbidden
	}
	return access, nil
}
