package iam

import (
	"encoding/json"
	"net/http"
)

type lifecycleCredentialDocument struct {
	TenantSlug string          `json:"tenant_slug"`
	Token      json.RawMessage `json:"token"`
}

func (handler *LifecycleHTTP) verifyIdentifier(w http.ResponseWriter, r *http.Request) {
	document, err := decodeIdentityDocument[lifecycleCredentialDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	token, err := decodeLoginPassword(document.Token)
	clear(document.Token)
	if err != nil || document.TenantSlug == "" {
		clear(token)
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	defer clear(token)
	err = handler.application.VerifyIdentifier(r.Context(),
		LifecycleCredentialInput{TenantSlug: document.TenantSlug, Token: token})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
