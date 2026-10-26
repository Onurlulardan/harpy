package httpwire

import (
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type Failure uint8

const (
	FailureRequestInvalid Failure = iota + 1
	FailureRouteNotFound
	FailureMethodNotAllowed
	FailureBodyTooLarge
	FailureMediaTypeUnsupported
	FailureRequestTimeout
	FailureInternal
	FailureServiceUnavailable
)

type failureDefinition struct {
	status  int
	code    string
	message string
}

func WriteFailure(
	w http.ResponseWriter,
	correlationID requestmeta.CorrelationID,
	failure Failure,
	headOnly bool,
) error {
	definition, exists := defineFailure(failure)
	if !exists {
		return fmt.Errorf("HTTP failure is invalid")
	}
	envelope, err := runtimev1.NewErrorEnvelope(
		definition.code,
		definition.message,
		correlationID.String(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("build HTTP failure: %w", err)
	}
	return WriteJSON(w, definition.status, correlationID, envelope, headOnly)
}

func defineFailure(failure Failure) (failureDefinition, bool) {
	switch failure {
	case FailureRequestInvalid:
		return failureDefinition{http.StatusBadRequest, "REQUEST_INVALID", "Request is invalid"}, true
	case FailureRouteNotFound:
		return failureDefinition{http.StatusNotFound, "ROUTE_NOT_FOUND", "Route not found"}, true
	case FailureMethodNotAllowed:
		return failureDefinition{http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"}, true
	case FailureBodyTooLarge:
		return failureDefinition{http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE", "Request body is too large"}, true
	case FailureMediaTypeUnsupported:
		return failureDefinition{http.StatusUnsupportedMediaType, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"}, true
	case FailureRequestTimeout:
		return failureDefinition{http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timed out"}, true
	case FailureInternal:
		return failureDefinition{http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"}, true
	case FailureServiceUnavailable:
		return failureDefinition{http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service is unavailable"}, true
	default:
		return failureDefinition{}, false
	}
}
