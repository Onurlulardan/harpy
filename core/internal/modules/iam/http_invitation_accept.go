package iam

import (
	"encoding/json"
	"net/http"
)

type invitationAcceptDocument struct {
	Token            json.RawMessage `json:"token"`
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Password         json.RawMessage `json:"password"`
	TenantSlug       string          `json:"tenant_slug"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *InvitationHTTP) accept(w http.ResponseWriter, r *http.Request) {
	document, err := decodeIdentityDocument[invitationAcceptDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	token, tokenErr := decodeLoginPassword(document.Token)
	password, passwordErr := decodeLoginPassword(document.Password)
	clear(document.Token)
	clear(document.Password)
	if tokenErr != nil || passwordErr != nil ||
		!validLoginText(document.FirstName, 200) ||
		!validLoginText(document.LastName, 200) ||
		!validLoginText(document.TenantSlug, 128) {
		clear(token)
		clear(password)
		handler.writeFailure(w, r, ErrInvitationHTTPInvalid)
		return
	}
	input := InvitationAcceptInput{
		Token: token, TenantSlug: document.TenantSlug,
		FirstName: document.FirstName, LastName: document.LastName,
		Password:         password,
		CustomAttributes: append([]byte(nil), document.CustomAttributes...),
	}
	defer input.Destroy()
	view, err := handler.application.Accept(r.Context(), input)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusCreated, view)
}
