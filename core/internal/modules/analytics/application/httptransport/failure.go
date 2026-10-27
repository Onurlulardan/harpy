package httptransport

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type failureDefinition struct {
	status        int
	code, message string
}

func (runtime *Runtime) Failure(writer http.ResponseWriter, request *http.Request, err error) {
	definition := runtime.failureFor(err)
	if definition.status >= 500 {
		runtime.Report(errors.New("analytics HTTP operation failed"))
	}
	correlation := runtime.Correlation(request)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		definition.code, definition.message, correlation, nil)
	if envelopeErr == nil {
		envelopeErr = writeJSON(writer, definition.status, correlation, envelope)
	}
	if envelopeErr != nil {
		runtime.Report(errors.New("analytics HTTP failure write failed"))
	}
}

func (runtime *Runtime) failureFor(err error) failureDefinition {
	switch {
	case errors.Is(err, errBodyTooLarge):
		return failure(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, errMediaUnsupported):
		return failure(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, ErrAuthentication):
		return failure(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, ErrForbidden):
		return failure(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, ErrNotFound):
		return failure(404, "ANALYTICS_NOT_FOUND", "Analytics resource was not found")
	case errors.Is(err, ErrConflict):
		return failure(409, "ANALYTICS_CONFLICT", "Analytics state conflicts")
	case errors.Is(err, ErrRateLimited):
		return failure(429, "ANALYTICS_RATE_LIMITED", "Public request rate limit was exceeded")
	case errors.Is(err, ErrInvalid), errors.Is(err, errDocumentInvalid):
		return failure(422, "ANALYTICS_INPUT_INVALID", "Analytics input is invalid")
	}
	if runtime != nil && runtime.failure != nil {
		status, code, message := runtime.failure(err)
		if status >= 400 && status <= 599 && code != "" && message != "" {
			return failure(status, code, message)
		}
	}
	return failure(503, "ANALYTICS_UNAVAILABLE", "Analytics is unavailable")
}

func failure(status int, code, message string) failureDefinition {
	return failureDefinition{status: status, code: code, message: message}
}
