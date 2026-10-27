package iam

import "net/http"

func (handler *LifecycleHTTP) revokeAllSessions(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	key, keyValid := identityIdempotencyKey(r)
	if !keyValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	if err := handler.application.RevokeAllSessions(r.Context(), claims, key); err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
