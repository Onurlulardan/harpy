package tenancy

import (
	"net/http"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

func (handler *TenantSettingsHTTP) accessFor(r *http.Request,
	permission TenantSettingsPermission, operation string,
) (TenantSettingsAccess, error) {
	if !handler.Ready() || r == nil {
		return TenantSettingsAccess{}, settings.ErrAuthentication
	}
	access, err := handler.access.Verify(r.Context(), TenantSettingsAccessRequest{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
		Permission:          permission, Operation: operation})
	if err != nil {
		return TenantSettingsAccess{}, err
	}
	if !access.validFor(permission, operation) {
		return TenantSettingsAccess{}, settings.ErrForbidden
	}
	return access, nil
}
