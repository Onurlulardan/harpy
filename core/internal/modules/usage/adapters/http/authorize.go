package usagehttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (handler *Handler) accessFor(request *http.Request, scope domain.Scope,
	operation string,
) (application.Access, error) {
	if !handler.Ready() || request == nil {
		return application.Access{}, domain.ErrUnavailable
	}
	permission := application.PermissionTenantRead
	if scope == domain.PlatformScope {
		permission = application.PermissionPlatformRead
	}
	access, err := handler.dependencies.Access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), request.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), request.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), request.Header.Values("X-Tenant-Slug")...),
		Scope:               scope, Permission: permission, Operation: operation,
	})
	if err != nil {
		return application.Access{}, err
	}
	if !access.Allows(scope, operation) || access.Permission() != permission {
		return application.Access{}, domain.ErrForbidden
	}
	return access, nil
}
