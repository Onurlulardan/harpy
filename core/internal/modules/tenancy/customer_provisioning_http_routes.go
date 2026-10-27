package tenancy

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

const customerProvisioningOperationsPath = "/api/platform/tenant-provisioning-operations/"

func (handler *CustomerProvisioningHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/platform/tenants",
			Handler: http.HandlerFunc(handler.create)},
		{Method: http.MethodPost, Pattern: "/api/admin/tenants",
			Handler: http.HandlerFunc(handler.create)},
		{Method: http.MethodPost, Pattern: "/api/auth/register-tenant",
			Handler: http.HandlerFunc(handler.legacyAnonymousRegistration)},
		{Method: http.MethodGet, Pattern: customerProvisioningOperationsPath + "{operation_id}",
			Handler: http.HandlerFunc(handler.read)},
		{Method: http.MethodPost,
			Pattern: "/api/platform/tenants/{tenant_id}/admin-setup/reissue",
			Handler: http.HandlerFunc(handler.reissueInitialAdminSetup)},
	}
}
