package tenancy

import (
	"net/http"
)

func (handler *CustomerLifecycleHTTP) legacyStats(w http.ResponseWriter, r *http.Request) {
	if _, valid := customerLifecycleTenantID(r); !valid || r.URL.RawQuery != "" ||
		!emptyRequestBody(r) {
		handler.writeCustomerLifecycleFailure(w, r, errCustomerLifecycleStatsMoved)
		return
	}
	w.Header().Set("Link", "</api/platform/usage>; rel=\"successor-version\", "+
		"</api/platform/billing>; rel=\"successor-version\"")
	handler.writeCustomerLifecycleFailure(w, r, errCustomerLifecycleStatsMoved)
}

func (handler *CustomerLifecycleHTTP) legacySwitchTenant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Link", "</api/auth/session>; rel=\"successor-version\"")
	w.Header().Set("Cache-Control", "no-store")
	handler.writeCustomerLifecycleFailure(w, r, errCustomerLifecycleSwitchGone)
}
