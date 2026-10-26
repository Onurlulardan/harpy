package tenancy

import (
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *CustomerProvisioningHTTP) legacyAnonymousRegistration(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Link", "</api/platform/tenants>; rel=\"successor-version\"")
	w.Header().Set("Cache-Control", "no-store")
	correlation := customerProvisioningCorrelation(r)
	envelope, err := runtimev1.NewErrorEnvelope(
		"LEGACY_ANONYMOUS_TENANT_REGISTRATION_RETIRED",
		"Anonymous tenant registration is retired; use authenticated platform provisioning",
		correlation.String(), nil,
	)
	if err == nil {
		err = httpwire.WriteJSON(w, http.StatusGone, correlation, envelope, false)
	}
	if err != nil {
		handler.report(err)
	}
}
