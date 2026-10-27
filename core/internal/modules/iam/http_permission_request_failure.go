package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *PermissionRequestHTTP) writeInputFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	failure := identityFailureInvalid
	if errors.Is(err, httpinput.ErrBodyTooLarge) {
		failure = identityFailureBodyTooLarge
	} else if errors.Is(err, httpinput.ErrMediaTypeUnsupported) {
		failure = identityFailureMediaType
	}
	handler.writeFailure(w, r, failure)
}

func (handler *PermissionRequestHTTP) writeApplicationFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, ErrPermissionRequestHTTPInvalid):
		handler.writeFailure(w, r, identityFailureInvalid)
	case errors.Is(err, ErrPermissionRequestHTTPForbidden):
		handler.writeFailure(w, r, identityFailureForbidden)
	case errors.Is(err, ErrPermissionRequestHTTPNotFound):
		handler.writeFailure(w, r, identityFailureNotFound)
	case errors.Is(err, ErrPermissionRequestHTTPConflict):
		handler.writeFailure(w, r, identityFailureConflict)
	default:
		handler.report(errors.New("IAM permission request application failed"))
		handler.writeFailure(w, r, identityFailureUnavailable)
	}
}

func (handler *PermissionRequestHTTP) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	failure identityHTTPFailure,
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
		handler.report(fmt.Errorf("write IAM permission request failure: %w", err))
	}
}
