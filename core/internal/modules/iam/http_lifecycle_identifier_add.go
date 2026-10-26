package iam

import "net/http"

type lifecycleIdentifierAddDocument struct {
	Email string `json:"email"`
}

func (handler *LifecycleHTTP) addIdentifier(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	document, err := decodeIdentityDocument[lifecycleIdentifierAddDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	view, err := handler.application.AddIdentifier(r.Context(), claims,
		LifecycleIdentifierAddInput{IdempotencyKey: key, Email: document.Email})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusCreated, view)
}
