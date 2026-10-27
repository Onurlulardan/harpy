package iam

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *InvitationHTTP) writeFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	status, code, message := invitationFailure(err)
	if status >= 500 {
		handler.report(errors.New("IAM invitation HTTP operation failed"))
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
		handler.report(errors.New("IAM invitation HTTP failure write failed"))
	}
}

func invitationFailure(err error) (int, string, string) {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return 415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"
	case errors.Is(err, errInvitationAuthentication):
		return 401, "AUTHENTICATION_FAILED", "Authentication failed"
	case errors.Is(err, ErrInvitationHTTPForbidden):
		return 403, "AUTHORIZATION_DENIED", "Operation is forbidden"
	case errors.Is(err, ErrInvitationHTTPInvalidCredential),
		errors.Is(err, ErrInvitationHTTPNotFound):
		return 404, "INVITATION_INVALID", "Invitation is invalid or unavailable"
	case errors.Is(err, ErrInvitationHTTPConflict):
		return 409, "INVITATION_CONFLICT", "Invitation operation conflicts"
	case errors.Is(err, ErrInvitationHTTPRecoveryRequired):
		return 409, "PROFILE_RECOVERY_REQUIRED", "Profile recovery is required"
	case errors.Is(err, ErrInvitationHTTPInvalid), errors.Is(err, httpinput.ErrInvalid):
		return 400, "REQUEST_INVALID", "Request is invalid"
	default:
		return 503, "IAM_UNAVAILABLE", "IAM is unavailable"
	}
}

var errInvitationAuthentication = errors.New("IAM invitation authentication failed")
