package tenancy

import (
	"errors"
	"net/http"
	"strings"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type customerLifecycleSuccess struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *CustomerLifecycleHTTP) writeCustomerLifecycleJSON(w http.ResponseWriter,
	r *http.Request, status int, data any,
) {
	handler.writeCustomerLifecycleWire(w, r, status,
		customerLifecycleSuccess{Success: true, Data: data})
}

func (handler *CustomerLifecycleHTTP) writeCustomerLifecycleLegacyJSON(w http.ResponseWriter,
	r *http.Request, status int, data any,
) {
	handler.writeCustomerLifecycleWire(w, r, status, data)
}

func (handler *CustomerLifecycleHTTP) writeCustomerLifecycleWire(w http.ResponseWriter,
	r *http.Request, status int, data any,
) {
	correlation := customerLifecycleCorrelation(r)
	if err := httpwire.WriteJSON(w, status, correlation, data, false); err != nil {
		handler.report(errors.New("customer lifecycle HTTP response write failed"))
	}
}

func (handler *CustomerLifecycleHTTP) writeCustomerLifecycleTenantSuccess(w http.ResponseWriter,
	r *http.Request, tenant lifecycle.Tenant, message string,
) {
	if customerLifecycleLegacyRequest(r) {
		handler.writeCustomerLifecycleLegacyJSON(w, r, http.StatusOK,
			customerLifecycleLegacyMutationView{Message: message,
				Data: customerLifecycleTenantView(tenant)})
		return
	}
	handler.writeCustomerLifecycleJSON(w, r, http.StatusOK,
		customerLifecycleTenantView(tenant))
}

func (handler *CustomerLifecycleHTTP) writeCustomerLifecycleFailure(w http.ResponseWriter,
	r *http.Request, err error,
) {
	status, code, message := customerLifecycleFailureDetails(err)
	correlation := customerLifecycleCorrelation(r)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(code, message,
		correlation.String(), nil)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(w, status, correlation, envelope, false)
	}
	if status >= 500 || envelopeErr != nil {
		handler.report(errors.New("customer lifecycle HTTP operation failed"))
	}
}

func customerLifecycleFailureDetails(err error) (int, string, string) {
	switch {
	case errors.Is(err, errCustomerLifecyclePrecondition):
		return 428, "REVISION_PRECONDITION_REQUIRED", "If-Match revision is required"
	case errors.Is(err, errCustomerLifecycleStatusMoved):
		return 426, "TENANT_STATUS_UPDATE_REDESIGNED", "Use the suspend or reactivate action"
	case errors.Is(err, errCustomerLifecycleStatsMoved):
		return 426, "TENANT_STATS_REDESIGNED", "Use bounded owner projections for tenant statistics"
	case errors.Is(err, errCustomerLifecycleSwitchGone):
		return 410, "TENANT_IMPERSONATION_RETIRED", "Platform operators cannot assume customer authority"
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, lifecycle.ErrAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, lifecycle.ErrModeDenied), errors.Is(err, lifecycle.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, lifecycle.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found"
	case errors.Is(err, lifecycle.ErrRevisionConflict):
		return 412, "REVISION_CONFLICT", "Tenant revision changed"
	case errors.Is(err, lifecycle.ErrIdempotencyConflict):
		return 409, "IDEMPOTENCY_KEY_REUSED", "Idempotency key was reused"
	case errors.Is(err, lifecycle.ErrStateConflict):
		return 409, "LIFECYCLE_STATE_CONFLICT", "Tenant lifecycle state conflicts"
	case errors.Is(err, lifecycle.ErrPurgeProof):
		return 409, "PURGE_PROOF_UNAVAILABLE", "Verified purge proof is unavailable"
	case errors.Is(err, lifecycle.ErrInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "TENANT_LIFECYCLE_UNAVAILABLE", "Tenant lifecycle is unavailable"
	}
}

func customerLifecycleLegacyRequest(r *http.Request) bool {
	return r != nil && strings.HasPrefix(r.URL.Path, "/api/admin/")
}

func customerLifecycleCorrelation(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
