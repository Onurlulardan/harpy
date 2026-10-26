package workhttp

import (
	"encoding/json"
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

var (
	errRequestInvalid   = errors.New("Work HTTP request is invalid")
	errBodyTooLarge     = errors.New("Work HTTP request body is too large")
	errMediaUnsupported = errors.New("Work HTTP media type is unsupported")
)

type failureDefinition struct {
	status        int
	code, message string
	issues        []runtimev1.Issue
}

func (handler *Handler) writeFailure(writer http.ResponseWriter, request *http.Request, err error) {
	definition := failure(err)
	if definition.status >= 500 {
		handler.report(errors.New("Work HTTP operation failed"))
	}
	correlation := handler.correlation(request)
	envelope, envelopeErr := runtimev1.NewErrorEnvelope(
		definition.code, definition.message, correlation, definition.issues)
	if envelopeErr == nil {
		var body []byte
		body, envelopeErr = json.Marshal(envelope)
		if envelopeErr == nil {
			envelopeErr = writeJSON(writer, definition.status, correlation, body)
		}
	}
	if envelopeErr != nil {
		handler.report(errors.New("Work HTTP failure write failed"))
	}
}

func failure(err error) failureDefinition {
	var queryError restapi.QueryError
	if errors.As(err, &queryError) {
		if status, valid := queryError.HTTPStatus(); valid {
			return failureDefinition{status: status, code: string(queryError.Code()),
				message: "REST Query request was rejected", issues: queryError.Issues()}
		}
	}
	switch {
	case errors.Is(err, errBodyTooLarge):
		return fail(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, errMediaUnsupported):
		return fail(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, ErrAuthentication):
		return fail(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, ErrForbidden), errors.Is(err, ports.ErrForbidden):
		return fail(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, ErrNotFound), errors.Is(err, ports.ErrNotFound):
		return fail(404, "WORK_NOT_FOUND", "Work resource was not found")
	case errors.Is(err, ErrConflict), errors.Is(err, ports.ErrConflict):
		return fail(409, "WORK_REVISION_CONFLICT", "Work state conflicts")
	case errors.Is(err, ErrInvalid), errors.Is(err, ports.ErrInvalid):
		return fail(422, "WORK_INPUT_INVALID", "Work input is invalid")
	case errors.Is(err, errRequestInvalid):
		return fail(400, "REQUEST_INVALID", "Request is invalid")
	default:
		return fail(503, "WORK_UNAVAILABLE", "Work is unavailable")
	}
}

func fail(status int, code, message string) failureDefinition {
	return failureDefinition{status: status, code: code, message: message}
}
