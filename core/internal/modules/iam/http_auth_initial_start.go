package iam

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type authHTTPInitialPasswordResponse struct {
	PasswordChangeRequired bool      `json:"password_change_required"`
	ExpiresAt              time.Time `json:"expires_at"`
}

func (handler *AuthHTTP) executeInitialPasswordStart(
	w http.ResponseWriter,
	r *http.Request,
	input AuthHTTPLoginInput,
	origin string,
) bool {
	copyInput := input
	copyInput.Password = append([]byte(nil), input.Password...)
	challenge, err := handler.application.StartInitialPassword(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), copyInput,
	)
	clear(copyInput.Password)
	if err != nil {
		if errors.Is(err, ErrAuthHTTPInvalidCredentials) {
			return false
		}
		handler.writeApplicationFailure(w, r, err)
		return true
	}
	if challenge == nil {
		handler.report(errors.New("IAM initial password returned no challenge"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return true
	}
	defer challenge.Destroy()
	now, valid := handler.authHTTPNow()
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return true
	}
	cookies, err := initialPasswordCookieValues(challenge, now)
	if err != nil {
		handler.report(errors.New("IAM initial password challenge presentation failed"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return true
	}
	appendSetCookies(w.Header(), expiredAuthCookieValues())
	appendSetCookies(w.Header(), cookies)
	response := authHTTPInitialPasswordResponse{
		PasswordChangeRequired: true,
		ExpiresAt:              challenge.ExpiresAt().UTC().Truncate(time.Second),
	}
	if err := httpwire.WriteJSON(
		w, http.StatusOK, authHTTPCorrelationID(r), response, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM initial password response: %w", err))
	}
	return true
}
