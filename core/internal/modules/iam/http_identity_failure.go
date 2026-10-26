package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type identityHTTPFailure uint8

const (
	identityFailureInvalid identityHTTPFailure = iota + 1
	identityFailureBodyTooLarge
	identityFailureMediaType
	identityFailureAuthentication
	identityFailureForbidden
	identityFailureNotFound
	identityFailureConflict
	identityFailureRecoveryRequired
	identityFailurePrecondition
	identityFailureUnavailable
)

type identityFailureDefinition struct {
	status int
	code   string
	text   string
}

func (handler *IdentityHTTP) writeIdentityInputFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	failure := identityFailureInvalid
	if errors.Is(err, httpinput.ErrBodyTooLarge) {
		failure = identityFailureBodyTooLarge
	} else if errors.Is(err, httpinput.ErrMediaTypeUnsupported) {
		failure = identityFailureMediaType
	}
	handler.writeIdentityFailure(w, r, failure)
}

func (handler *IdentityHTTP) writeIdentityApplicationFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	switch {
	case errors.Is(err, ErrIdentityHTTPInvalid):
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
	case errors.Is(err, ErrIdentityHTTPForbidden):
		handler.writeIdentityFailure(w, r, identityFailureForbidden)
	case errors.Is(err, ErrIdentityHTTPNotFound):
		handler.writeIdentityFailure(w, r, identityFailureNotFound)
	case errors.Is(err, ErrIdentityHTTPConflict):
		handler.writeIdentityFailure(w, r, identityFailureConflict)
	case errors.Is(err, ErrIdentityHTTPRecoveryRequired):
		handler.writeIdentityFailure(w, r, identityFailureRecoveryRequired)
	default:
		handler.report(errors.New("IAM system-resource application failed"))
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
	}
}

func (handler *IdentityHTTP) writeQueryFailure(
	w http.ResponseWriter, r *http.Request, err error,
) {
	var queryError queryapp.Failure
	if errors.As(err, &queryError) {
		envelope, envelopeErr := runtimev1.NewErrorEnvelope(
			queryError.Code(), "Query is invalid",
			authHTTPCorrelationID(r).String(), queryError.Issues(),
		)
		if queryError.HTTPStatus() >= 400 && envelopeErr == nil {
			if writeErr := httpwire.WriteJSON(
				w, queryError.HTTPStatus(), authHTTPCorrelationID(r), envelope, false,
			); writeErr == nil {
				return
			}
		}
		handler.report(errors.New("IAM REST Query error contract failed"))
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeIdentityApplicationFailure(w, r, err)
}

func (handler *IdentityHTTP) writeIdentityFailure(
	w http.ResponseWriter, r *http.Request, failure identityHTTPFailure,
) {
	definition, found := identityFailure(failure)
	if !found {
		definition, _ = identityFailure(identityFailureUnavailable)
	}
	envelope, err := runtimev1.NewErrorEnvelope(
		definition.code, definition.text, authHTTPCorrelationID(r).String(), nil,
	)
	if err == nil {
		err = httpwire.WriteJSON(
			w, definition.status, authHTTPCorrelationID(r), envelope, false,
		)
	}
	if err != nil {
		handler.report(fmt.Errorf("write IAM system-resource failure: %w", err))
	}
}

func identityFailure(failure identityHTTPFailure) (identityFailureDefinition, bool) {
	definitions := map[identityHTTPFailure]identityFailureDefinition{
		identityFailureInvalid:          {400, "REQUEST_INVALID", "Request is invalid"},
		identityFailureBodyTooLarge:     {413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"},
		identityFailureMediaType:        {415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"},
		identityFailureAuthentication:   {401, "AUTHENTICATION_FAILED", "Authentication failed"},
		identityFailureForbidden:        {403, "AUTHORIZATION_DENIED", "Operation is forbidden"},
		identityFailureNotFound:         {404, "RESOURCE_NOT_FOUND", "Resource was not found"},
		identityFailureConflict:         {409, "RESOURCE_CONFLICT", "Resource operation conflicts"},
		identityFailureRecoveryRequired: {409, "PROFILE_RECOVERY_REQUIRED", "Profile recovery is required"},
		identityFailurePrecondition:     {428, "PRECONDITION_REQUIRED", "Revision precondition is required"},
		identityFailureUnavailable:      {503, "IAM_UNAVAILABLE", "IAM is unavailable"},
	}
	definition, found := definitions[failure]
	return definition, found
}
