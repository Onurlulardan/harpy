package iam

import (
	"encoding/json"
	"net/http"
)

type lifecyclePasswordResetCompleteDocument struct {
	TenantSlug string          `json:"tenant_slug"`
	Token      json.RawMessage `json:"token"`
	Password   json.RawMessage `json:"password"`
}

func (handler *LifecycleHTTP) completePasswordReset(w http.ResponseWriter, r *http.Request) {
	document, err := decodeIdentityDocument[lifecyclePasswordResetCompleteDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	token, tokenErr := decodeLoginPassword(document.Token)
	password, passwordErr := decodeLoginPassword(document.Password)
	clear(document.Token)
	clear(document.Password)
	if tokenErr != nil || passwordErr != nil || document.TenantSlug == "" {
		clear(token)
		clear(password)
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	defer clear(token)
	defer clear(password)
	err = handler.application.CompletePasswordReset(r.Context(),
		LifecyclePasswordResetCompleteInput{TenantSlug: document.TenantSlug,
			Token: token, Password: password})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
