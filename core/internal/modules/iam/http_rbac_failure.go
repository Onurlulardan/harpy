package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *RBACHTTP) writeInputFailure(
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

func (handler *RBACHTTP) writeApplicationFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, ErrRBACHTTPInvalid):
		handler.writeFailure(w, r, identityFailureInvalid)
	case errors.Is(err, ErrRBACHTTPForbidden):
		handler.writeFailure(w, r, identityFailureForbidden)
	case errors.Is(err, ErrRBACHTTPNotFound):
		handler.writeFailure(w, r, identityFailureNotFound)
	case errors.Is(err, ErrRBACHTTPConflict):
		handler.writeFailure(w, r, identityFailureConflict)
	default:
		handler.report(errors.New("IAM RBAC application failed"))
		handler.writeFailure(w, r, identityFailureUnavailable)
	}
}

func (handler *RBACHTTP) writeFailure(
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
		handler.report(fmt.Errorf("write IAM RBAC failure: %w", err))
	}
}

func (handler *RBACHTTP) cacheRetired(w http.ResponseWriter, r *http.Request) {
	if _, valid := handler.claims(r); !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	envelope, err := runtimev1.NewErrorEnvelope(
		"PERMISSION_CACHE_RETIRED",
		"Permission cache endpoints were retired; authorization is versioned and transactional",
		authHTTPCorrelationID(r).String(), nil,
	)
	if err == nil {
		err = httpwire.WriteJSON(
			w, http.StatusGone, authHTTPCorrelationID(r), envelope, false,
		)
	}
	if err != nil {
		handler.report(fmt.Errorf("write IAM RBAC cache retirement: %w", err))
	}
}
