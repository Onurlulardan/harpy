package iam

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

var errLifecycleAuthentication = errors.New("IAM lifecycle authentication failed")

func (handler *LifecycleHTTP) writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := lifecycleFailure(err)
	if status >= 500 {
		handler.report(errors.New("IAM lifecycle HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(code, message,
		authHTTPCorrelationID(r).String(), nil)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(w, status, authHTTPCorrelationID(r), envelope, false)
	}
	if envelopeErr != nil {
		handler.report(errors.New("IAM lifecycle failure write failed"))
	}
}

func lifecycleFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, errLifecycleAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrLifecycleHTTPForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, ErrLifecycleHTTPNotFound), errors.Is(err, ErrLifecycleHTTPInvalidCredential):
		return 404, "LIFECYCLE_CREDENTIAL_INVALID", "Credential is invalid or unavailable"
	case errors.Is(err, ErrLifecycleHTTPConflict):
		return 409, "LIFECYCLE_CONFLICT", "Lifecycle operation conflicts"
	case errors.Is(err, ErrLifecycleHTTPInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "IAM_UNAVAILABLE", "IAM is unavailable"
	}
}
