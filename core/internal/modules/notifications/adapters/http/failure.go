package notificationhttp

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
)

type failureDefinition struct {
	status        int
	code, message string
}

func (handler *Handler) writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	definition := failure(err)
	if definition.status >= 500 {
		handler.report(errors.New("notification HTTP operation failed"))
	}
	correlation := handler.correlation(r)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		definition.code, definition.message, correlation, nil)
	if envelopeErr == nil {
		envelopeErr = writeJSON(w, definition.status, correlation, envelope)
	}
	if envelopeErr != nil {
		handler.report(errors.New("notification HTTP failure write failed"))
	}
}

func failure(err error) failureDefinition {
	switch {
	case errors.Is(err, errBodyTooLarge):
		return errorDefinition(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, errMediaUnsupported):
		return errorDefinition(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, ErrAuthentication):
		return errorDefinition(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, ErrForbidden), errors.Is(err, application.ErrForbidden),
		errors.Is(err, api.ErrForbidden):
		return errorDefinition(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, ErrNotFound), errors.Is(err, application.ErrNotFound):
		return errorDefinition(404, "NOTIFICATION_NOT_FOUND", "Notification was not found")
	case errors.Is(err, api.ErrEffectKeyReused):
		return errorDefinition(409, "NOTIFICATION_EFFECT_KEY_REUSED", "Notification effect conflicts")
	case errors.Is(err, ErrConflict), errors.Is(err, application.ErrConflict):
		return errorDefinition(409, "NOTIFICATION_CONFLICT", "Notification state conflicts")
	case errors.Is(err, ErrInvalid), errors.Is(err, application.ErrInvalid),
		errors.Is(err, api.ErrIntentInvalid):
		return errorDefinition(422, "NOTIFICATION_INPUT_INVALID", "Notification input is invalid")
	case errors.Is(err, errRequestInvalid), errors.Is(err, errQueryInvalid):
		return errorDefinition(400, "REQUEST_INVALID", "Request is invalid")
	default:
		return errorDefinition(503, "NOTIFICATION_UNAVAILABLE", "Notification service is unavailable")
	}
}

func errorDefinition(status int, code, message string) failureDefinition {
	return failureDefinition{status: status, code: code, message: message}
}
