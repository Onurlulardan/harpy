package billing

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

var platformLicenseRequestInvalid = errors.New("platform license request is invalid")

func (handler *PlatformLicenseHTTPHandler) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	status, code, message := platformLicenseFailure(err)
	if status >= 500 {
		handler.report(errors.New("platform license HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		code, message, correlationID(r).String(), nil,
	)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(
			w, status, correlationID(r), envelope, false,
		)
	}
	if envelopeErr != nil {
		handler.report(errors.New("platform license HTTP failure write failed"))
	}
}

func platformLicenseFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrPlatformLicenseHTTPAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrPlatformLicenseHTTPForbidden),
		errors.Is(err, platformlicenseapp.ErrForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, platformlicenseapp.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found"
	case errors.Is(err, platformlicenseapp.ErrIdempotencyReused):
		return 409, "IDEMPOTENCY_KEY_REUSED", "Idempotency key was reused"
	case errors.Is(err, platformlicenseapp.ErrConflict):
		return 409, "PLATFORM_LICENSE_CONFLICT", "Platform license state conflicts"
	case errors.Is(err, platformlicenseapp.ErrInProgress):
		return 409, "IDEMPOTENCY_IN_PROGRESS", "Operation is in progress"
	case errors.Is(err, platformlicenseapp.ErrInvalid),
		errors.Is(err, platformLicenseRequestInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "PLATFORM_LICENSE_UNAVAILABLE", "Platform license is unavailable"
	}
}
