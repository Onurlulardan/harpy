package tenancy

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

const customerLifecycleOperationsPath = "/api/platform/tenant-lifecycle-operations/"

func (handler *CustomerLifecycleHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/platform/tenants", Handler: http.HandlerFunc(handler.list)},
		{Method: http.MethodGet, Pattern: "/api/platform/tenants/{tenant_id}", Handler: http.HandlerFunc(handler.get)},
		{Method: http.MethodPut, Pattern: "/api/platform/tenants/{tenant_id}", Handler: http.HandlerFunc(handler.update)},
		{Method: http.MethodPost, Pattern: "/api/platform/tenants/{tenant_id}/suspend", Handler: http.HandlerFunc(handler.suspend)},
		{Method: http.MethodPost, Pattern: "/api/platform/tenants/{tenant_id}/reactivate", Handler: http.HandlerFunc(handler.reactivate)},
		{Method: http.MethodPost, Pattern: "/api/platform/tenants/{tenant_id}/offboard", Handler: http.HandlerFunc(handler.offboard)},
		{Method: http.MethodPost, Pattern: "/api/platform/tenants/{tenant_id}/purge", Handler: http.HandlerFunc(handler.purge)},
		{Method: http.MethodGet, Pattern: customerLifecycleOperationsPath + "{operation_id}", Handler: http.HandlerFunc(handler.readOperation)},
		{Method: http.MethodGet, Pattern: "/api/admin/tenants", Handler: http.HandlerFunc(handler.list)},
		{Method: http.MethodGet, Pattern: "/api/admin/tenants/{tenant_id}", Handler: http.HandlerFunc(handler.get)},
		{Method: http.MethodPut, Pattern: "/api/admin/tenants/{tenant_id}", Handler: http.HandlerFunc(handler.legacyUpdate)},
		{Method: http.MethodPost, Pattern: "/api/admin/tenants/{tenant_id}/suspend", Handler: http.HandlerFunc(handler.suspend)},
		{Method: http.MethodPost, Pattern: "/api/admin/tenants/{tenant_id}/reactivate", Handler: http.HandlerFunc(handler.reactivate)},
		{Method: http.MethodDelete, Pattern: "/api/admin/tenants/{tenant_id}", Handler: http.HandlerFunc(handler.offboard)},
		{Method: http.MethodGet, Pattern: "/api/admin/tenants/{tenant_id}/stats", Handler: http.HandlerFunc(handler.legacyStats)},
		{Method: http.MethodPost, Pattern: "/api/admin/switch-tenant", Handler: http.HandlerFunc(handler.legacySwitchTenant)},
	}
}
