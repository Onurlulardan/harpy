package iam

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

var errProfileAuthentication = errors.New("IAM profile authentication failed")

func (handler *ProfileHTTP) writeFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	status, code, message := profileHTTPFailure(err)
	if status >= 500 {
		handler.report(errors.New("IAM profile HTTP operation failed"))
	}
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		code, message, authHTTPCorrelationID(r).String(), nil,
	)
	if envelopeErr == nil {
		envelopeErr = httpwire.WriteJSON(
			w, status, authHTTPCorrelationID(r), envelope, false,
		)
	}
	if envelopeErr != nil {
		handler.report(errors.New("IAM profile HTTP failure write failed"))
	}
}

func profileHTTPFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, errProfileAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrProfileHTTPFieldForbidden):
		return 422, "PROFILE_FIELD_FORBIDDEN", "Profile field cannot be changed"
	case errors.Is(err, ErrProfileHTTPForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, ErrProfileHTTPNotFound):
		return 404, "RESOURCE_NOT_FOUND", "Resource was not found"
	case errors.Is(err, ErrProfileHTTPConflict):
		return 409, "RESOURCE_CONFLICT", "Resource operation conflicts"
	case errors.Is(err, ErrProfileHTTPRecoveryRequired):
		return 409, "PROFILE_RECOVERY_REQUIRED", "Profile recovery is required"
	case errors.Is(err, ErrProfileHTTPRecoveryBudgetExceeded):
		return 409, "PROFILE_RECOVERY_BUDGET_EXCEEDED", "Profile recovery budget is exceeded"
	case errors.Is(err, ErrProfileHTTPInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "IAM_UNAVAILABLE", "IAM is unavailable"
	}
}
