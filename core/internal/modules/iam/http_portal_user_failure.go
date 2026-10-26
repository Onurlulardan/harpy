package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type portalUserFailure uint8

const (
	portalUserFailureInvalid portalUserFailure = iota + 1
	portalUserFailureBodyTooLarge
	portalUserFailureMediaType
	portalUserFailureAuthentication
	portalUserFailureForbidden
	portalUserFailureNotFound
	portalUserFailureConflict
	portalUserFailurePrecondition
	portalUserFailureUnavailable
)

type portalUserFailureDefinition struct {
	status        int
	code, message string
}

func (handler *PortalUserHTTP) writeInputFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	failure := portalUserFailureInvalid
	if errors.Is(err, httpinput.ErrBodyTooLarge) {
		failure = portalUserFailureBodyTooLarge
	}
	if errors.Is(err, httpinput.ErrMediaTypeUnsupported) {
		failure = portalUserFailureMediaType
	}
	handler.writeFailure(w, r, failure)
}

func (handler *PortalUserHTTP) writeApplicationFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	switch {
	case errors.Is(err, ErrPortalUserHTTPInvalid):
		handler.writeFailure(w, r, portalUserFailureInvalid)
	case errors.Is(err, ErrPortalUserHTTPForbidden):
		handler.writeFailure(w, r, portalUserFailureForbidden)
	case errors.Is(err, ErrPortalUserHTTPNotFound):
		handler.writeFailure(w, r, portalUserFailureNotFound)
	case errors.Is(err, ErrPortalUserHTTPConflict):
		handler.writeFailure(w, r, portalUserFailureConflict)
	default:
		handler.report(errors.New("IAM portal user application failed"))
		handler.writeFailure(w, r, portalUserFailureUnavailable)
	}
}

func (handler *PortalUserHTTP) writeFailure(
	w http.ResponseWriter, r *http.Request, failure portalUserFailure,
) {
	definitions := map[portalUserFailure]portalUserFailureDefinition{
		portalUserFailureInvalid:        {400, "REQUEST_INVALID", "Request is invalid"},
		portalUserFailureBodyTooLarge:   {413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"},
		portalUserFailureMediaType:      {415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"},
		portalUserFailureAuthentication: {401, "AUTHENTICATION_FAILED", "Authentication failed"},
		portalUserFailureForbidden:      {403, "AUTHORIZATION_DENIED", "Operation is forbidden"},
		portalUserFailureNotFound:       {404, "RESOURCE_NOT_FOUND", "Portal user was not found"},
		portalUserFailureConflict:       {409, "RESOURCE_CONFLICT", "Portal user operation conflicts"},
		portalUserFailurePrecondition:   {428, "PRECONDITION_REQUIRED", "Revision precondition is required"},
		portalUserFailureUnavailable:    {503, "IAM_UNAVAILABLE", "IAM is unavailable"},
	}
	definition, found := definitions[failure]
	if !found {
		definition = definitions[portalUserFailureUnavailable]
	}
	envelope, err := runtimev1.NewErrorEnvelope(definition.code, definition.message,
		authHTTPCorrelationID(r).String(), nil)
	if err == nil {
		err = httpwire.WriteJSON(w, definition.status, authHTTPCorrelationID(r), envelope, false)
	}
	if err != nil {
		handler.report(fmt.Errorf("write IAM portal user failure: %w", err))
	}
}
