package iam

import (
	"encoding/json"
	"net/http"
)

type lifecyclePasswordChangeDocument struct {
	CurrentPassword json.RawMessage `json:"current_password"`
	NewPassword     json.RawMessage `json:"new_password"`
}

func (handler *LifecycleHTTP) changePassword(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	document, err := decodeIdentityDocument[lifecyclePasswordChangeDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	current, currentErr := decodeLoginPassword(document.CurrentPassword)
	next, nextErr := decodeLoginPassword(document.NewPassword)
	clear(document.CurrentPassword)
	clear(document.NewPassword)
	if !keyValid || !revisionValid || currentErr != nil || nextErr != nil {
		clear(current)
		clear(next)
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	defer clear(current)
	defer clear(next)
	err = handler.application.ChangePassword(r.Context(), claims,
		LifecyclePasswordChangeInput{IdempotencyKey: key, ExpectedRevision: revision,
			CurrentPassword: current, NewPassword: next})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
