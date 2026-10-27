package iam

import "net/http"

type lifecyclePasswordResetStartDocument struct {
	TenantSlug     string `json:"tenant_slug"`
	IdentifierKind string `json:"identifier_kind"`
	Identifier     string `json:"identifier"`
}

func (handler *LifecycleHTTP) startPasswordReset(w http.ResponseWriter, r *http.Request) {
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	document, err := decodeIdentityDocument[lifecyclePasswordResetStartDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	err = handler.application.StartPasswordReset(r.Context(), LifecyclePasswordResetStartInput{
		TenantSlug: document.TenantSlug, IdentifierKind: document.IdentifierKind,
		Identifier: document.Identifier, IdempotencyKey: key})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusAccepted, map[string]bool{"accepted": true})
}
