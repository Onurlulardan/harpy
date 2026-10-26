package platformops

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *HTTPHandler) writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := platformOpsFailure(err)
	if status >= 500 {
		handler.report(errors.New("platform operations HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		code, message, platformOpsCorrelationID(r).String(), nil)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(
			w, status, platformOpsCorrelationID(r), envelope, false)
	}
	if envelopeErr != nil {
		handler.report(errors.New("platform operations HTTP failure write failed"))
	}
}

func platformOpsFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, ErrHTTPAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrHTTPForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, platformopsapi.ErrDenied):
		return 403, "RECOVERY_ANCHOR_DENIED", "Recovery anchor operation is forbidden"
	case errors.Is(err, platformopsapi.ErrNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found"
	case errors.Is(err, platformopsapi.ErrConflict):
		return 409, "RECOVERY_STATE_CONFLICT", "Recovery state conflicts"
	case errors.Is(err, platformopsapi.ErrInvalid), errors.Is(err, httpinput.ErrInvalid),
		errors.Is(err, errHTTPRequestInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "PLATFORM_OPERATIONS_UNAVAILABLE", "Platform operations are unavailable"
	}
}
