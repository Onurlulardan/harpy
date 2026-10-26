package iam

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) startReauthentication(w http.ResponseWriter, r *http.Request) {
	origin, valid := exactOrigin(r)
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	input, err := decodeReauthenticationStart(w, r, origin)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	claims, valid := handler.reauthenticationClaims(w, r)
	if !valid {
		return
	}
	challenge, err := handler.application.StartReauthentication(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	if challenge == nil {
		handler.report(errors.New("IAM reauthentication returned no challenge"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer challenge.Destroy()
	response, err := reauthenticationResponse(challenge)
	if err != nil {
		handler.report(errors.New("IAM reauthentication challenge could not be presented"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	if err := httpwire.WriteJSON(w, http.StatusCreated,
		authHTTPCorrelationID(r), response, false); err != nil {
		handler.report(fmt.Errorf("write IAM reauthentication challenge: %w", err))
	}
}
