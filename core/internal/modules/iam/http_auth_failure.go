package iam

import (
	"errors"
	"fmt"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type authHTTPFailure uint8

const (
	authFailureRequestInvalid authHTTPFailure = iota + 1
	authFailureBodyTooLarge
	authFailureMediaTypeUnsupported
	authFailureAuthentication
	authFailurePasswordRejected
	authFailureCSRF
	authFailureUnavailable
)

type authHTTPFailureDefinition struct {
	status  int
	code    string
	message string
}

func (handler *AuthHTTP) writeInputFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	failure := authFailureRequestInvalid
	if errors.Is(err, httpinput.ErrBodyTooLarge) {
		failure = authFailureBodyTooLarge
	} else if errors.Is(err, httpinput.ErrMediaTypeUnsupported) {
		failure = authFailureMediaTypeUnsupported
	}
	handler.writeAuthFailure(w, r, failure)
}

func (handler *AuthHTTP) writeApplicationFailure(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if errors.Is(err, ErrAuthHTTPPasswordRejected) {
		handler.writeAuthFailure(w, r, authFailurePasswordRejected)
		return
	}
	if authHTTPRejected(err) {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return
	}
	handler.report(errors.New("IAM browser authentication operation failed"))
	handler.writeAuthFailure(w, r, authFailureUnavailable)
}

func authHTTPRejected(err error) bool {
	return errors.Is(err, ErrAuthHTTPInvalidCredentials) ||
		errors.Is(err, ErrAuthHTTPInvalidSession) ||
		errors.Is(err, authentication.ErrInvalidCredentials)
}

func (handler *AuthHTTP) writeAuthFailure(
	w http.ResponseWriter,
	r *http.Request,
	failure authHTTPFailure,
) {
	definition, found := defineAuthHTTPFailure(failure)
	if !found {
		handler.report(errors.New("IAM auth HTTP failure is invalid"))
		definition, _ = defineAuthHTTPFailure(authFailureUnavailable)
	}
	correlationID := authHTTPCorrelationID(r)
	envelope, err := runtimev1.NewErrorEnvelope(
		definition.code, definition.message, correlationID.String(), nil,
	)
	if err == nil {
		err = httpwire.WriteJSON(w, definition.status, correlationID, envelope, false)
	}
	if err != nil {
		handler.report(fmt.Errorf("write IAM auth HTTP failure: %w", err))
	}
}

func defineAuthHTTPFailure(failure authHTTPFailure) (authHTTPFailureDefinition, bool) {
	switch failure {
	case authFailureRequestInvalid:
		return authHTTPFailureDefinition{400, "REQUEST_INVALID", "Request is invalid"}, true
	case authFailureBodyTooLarge:
		return authHTTPFailureDefinition{413, "REQUEST_BODY_TOO_LARGE", "Request body is too large"}, true
	case authFailureMediaTypeUnsupported:
		return authHTTPFailureDefinition{415, "MEDIA_TYPE_UNSUPPORTED", "Media type is unsupported"}, true
	case authFailureAuthentication:
		return authHTTPFailureDefinition{401, "AUTHENTICATION_FAILED", "Authentication failed"}, true
	case authFailurePasswordRejected:
		return authHTTPFailureDefinition{400, "PASSWORD_POLICY_REJECTED", "Password does not meet policy"}, true
	case authFailureCSRF:
		return authHTTPFailureDefinition{403, "AUTH_CSRF_INVALID", "Request proof is invalid"}, true
	case authFailureUnavailable:
		return authHTTPFailureDefinition{503, "AUTHENTICATION_UNAVAILABLE", "Authentication is unavailable"}, true
	default:
		return authHTTPFailureDefinition{}, false
	}
}

func authHTTPCorrelationID(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
