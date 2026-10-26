package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
)

func (handler *Handler) accessFor(request *http.Request, scope application.AccessScope,
	permission, operation string,
) (application.Access, error) {
	if handler == nil || request == nil {
		return application.Access{}, application.ErrUnavailable
	}
	access, err := handler.access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), request.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), request.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), request.Header.Values("X-Tenant-Slug")...),
		Scope:               scope, Permission: permission, Operation: operation,
	})
	if err != nil || !access.Allows(scope, permission, operation) {
		if err != nil {
			return application.Access{}, err
		}
		return application.Access{}, application.ErrForbidden
	}
	return access, nil
}
