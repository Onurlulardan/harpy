package api

import (
	"errors"
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

var (
	errLayoutHTTPRequestInvalid  = errors.New("system resource layout HTTP request is invalid")
	errLayoutHTTPPrecondition    = errors.New("system resource layout HTTP precondition is required")
	errLayoutHTTPResourceUnknown = errors.New("system resource layout HTTP resource is unknown")
)

type layoutHTTPFailureDefinition struct {
	status  int
	code    string
	message string
}

type layoutHTTPErrorEnvelope struct {
	Success bool                 `json:"success"`
	Error   layoutHTTPErrorEntry `json:"error"`
}

type layoutHTTPErrorEntry struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	Issues        []any  `json:"issues"`
	CorrelationID string `json:"correlation_id"`
}

func (handler *systemResourceLayoutHTTPHandler) writeHTTPFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	definition := layoutHTTPFailure(err)
	if definition.status >= 500 {
		handler.report(errors.New("system resource layout HTTP operation failed"))
	}
	envelope := layoutHTTPErrorEnvelope{Error: layoutHTTPErrorEntry{
		Code: definition.code, Message: definition.message, Issues: []any{},
		CorrelationID: layoutHTTPCorrelationID(r).String(),
	}}
	envelopeErr := httpwire.WriteJSON(
		w, definition.status, layoutHTTPCorrelationID(r), envelope, false,
	)
	if envelopeErr != nil {
		handler.report(errors.New("system resource layout HTTP failure write failed"))
	}
}

func layoutHTTPFailure(err error) layoutHTTPFailureDefinition {
	switch {
	case errors.Is(err, httpinput.ErrBodyTooLarge):
		return layoutHTTPError(413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
	case errors.Is(err, httpinput.ErrMediaTypeUnsupported):
		return layoutHTTPError(415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported")
	case errors.Is(err, layouts.ErrLayoutAccessAuthentication):
		return layoutHTTPError(401, "AUTHENTICATION_FAILED", "Authentication failed")
	case errors.Is(err, layouts.ErrLayoutAccessForbidden):
		return layoutHTTPError(403, "AUTHORIZATION_DENIED", "Operation is forbidden")
	case errors.Is(err, errLayoutHTTPResourceUnknown):
		return layoutHTTPError(404, "SYSTEM_RESOURCE_UNKNOWN", "System resource was not found")
	case errors.Is(err, layouts.ErrLayoutNotFound):
		return layoutHTTPError(404, "SYSTEM_RESOURCE_LAYOUT_NOT_FOUND", "Layout revision was not found")
	case errors.Is(err, layouts.ErrLayoutConflict):
		return layoutHTTPError(409, "SYSTEM_RESOURCE_LAYOUT_CONFLICT", "Layout revision conflicts")
	case errors.Is(err, errLayoutHTTPPrecondition):
		return layoutHTTPError(428, "PRECONDITION_REQUIRED", "Layout precondition is required")
	case errors.Is(err, layouts.ErrLayoutInvalid), errors.Is(err, layouts.ErrContract):
		return layoutHTTPError(422, "SYSTEM_RESOURCE_LAYOUT_INVALID", "Layout input is invalid")
	case errors.Is(err, errLayoutHTTPRequestInvalid), errors.Is(err, httpinput.ErrInvalid):
		return layoutHTTPError(400, "REQUEST_INVALID", "Request is invalid")
	case errors.Is(err, layouts.ErrLayoutAccessUnavailable):
		return layoutHTTPError(503, "SYSTEM_RESOURCE_LAYOUT_AUTHORITY_UNAVAILABLE", "Layout authority is unavailable")
	default:
		return layoutHTTPError(503, "SYSTEM_RESOURCE_LAYOUT_UNAVAILABLE", "Layout service is unavailable")
	}
}

func layoutHTTPError(status int, code, message string) layoutHTTPFailureDefinition {
	return layoutHTTPFailureDefinition{status: status, code: code, message: message}
}

func layoutHTTPCorrelationID(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
