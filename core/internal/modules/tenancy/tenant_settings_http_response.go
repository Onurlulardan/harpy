package tenancy

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type tenantSettingsSuccess struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
}

func (handler *TenantSettingsHTTP) writeTenantSettingsSuccess(w http.ResponseWriter,
	r *http.Request, status int, message string, data any,
) {
	correlation := tenantSettingsCorrelation(r)
	if err := httpwire.WriteJSON(w, status, correlation, tenantSettingsSuccess{
		Success: true, Message: message, Data: data}, false); err != nil {
		handler.report(errors.New("tenant settings response write failed"))
	}
}

func (handler *TenantSettingsHTTP) writeTenantSettingsFailure(w http.ResponseWriter,
	r *http.Request, err error,
) {
	status, code, message := tenantSettingsFailureDetails(err)
	correlation := tenantSettingsCorrelation(r)
	envelope, writeErr := runtimev1.NewErrorEnvelope(code, message, correlation.String(), nil)
	if writeErr == nil {
		writeErr = httpwire.WriteJSON(w, status, correlation, envelope, false)
	}
	if status >= 500 || writeErr != nil {
		handler.report(errors.New("tenant settings operation failed"))
	}
}

func tenantSettingsFailureDetails(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, settings.ErrAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, settings.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, settings.ErrPrecondition):
		return 428, "REVISION_REQUIRED", "If-Match revision is required"
	case errors.Is(err, settings.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Tenant settings were not found"
	case errors.Is(err, settings.ErrConflict):
		return 409, "SETTINGS_CONFLICT", "Tenant settings revision or idempotency key conflicts"
	case errors.Is(err, settings.ErrInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "TENANT_SETTINGS_UNAVAILABLE", "Tenant settings are unavailable"
	}
}

func tenantSettingsCorrelation(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
