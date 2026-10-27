package iam

import (
	"net/http"
)

type invitationCreateDocument struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (handler *InvitationHTTP) create(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errInvitationAuthentication)
		return
	}
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, ErrInvitationHTTPInvalid)
		return
	}
	document, err := decodeIdentityDocument[invitationCreateDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	view, err := handler.application.Create(
		r.Context(), claims, InvitationCreateInput{
			IdempotencyKey: key, Email: document.Email, Role: document.Role,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeInvitation(w, r, http.StatusCreated, view)
}
