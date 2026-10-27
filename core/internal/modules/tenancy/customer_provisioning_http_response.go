package tenancy

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type customerProvisioningSuccess struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *CustomerProvisioningHTTP) writeCustomerProvisioningJSON(
	w http.ResponseWriter, r *http.Request, status int, data any,
) {
	correlation := customerProvisioningCorrelation(r)
	if err := httpwire.WriteJSON(w, status, correlation,
		customerProvisioningSuccess{Success: true, Data: data}, false); err != nil {
		handler.report(errors.New("customer provisioning HTTP response write failed"))
	}
}

func (handler *CustomerProvisioningHTTP) writeCustomerProvisioningFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	status, code, message := customerProvisioningFailureDetails(err)
	correlation := customerProvisioningCorrelation(r)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		code, message, correlation.String(), nil,
	)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(w, status, correlation, envelope, false)
	}
	if status >= 500 || envelopeErr != nil {
		handler.report(errors.New("customer provisioning HTTP operation failed"))
	}
}

func customerProvisioningFailureDetails(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, provisioning.ErrAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, provisioning.ErrModeDenied):
		return 403, "DEPLOYMENT_MODE_DENIED", "Operation is unavailable in this deployment mode"
	case errors.Is(err, provisioning.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, provisioning.ErrConflict):
		return 409, "IDEMPOTENCY_KEY_REUSED", "Idempotency key was reused"
	case errors.Is(err, provisioning.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found"
	case errors.Is(err, provisioning.ErrInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "TENANT_PROVISIONING_UNAVAILABLE", "Tenant provisioning is unavailable"
	}
}

func customerProvisioningCorrelation(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
