package iam

import (
	"encoding/json"
	"net/http"
)

func (handler *ProfileHTTP) updateSelf(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	key, valid := profileSelfIdempotencyKey(r)
	if !valid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	values, err := decodeIdentityDocument[map[string]json.RawMessage](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if !profileSelfKeysAllowed(values) {
		handler.writeFailure(w, r, ErrProfileHTTPFieldForbidden)
		return
	}
	view, err := handler.application.UpdateSelf(
		r.Context(), claims,
		ProfileSelfUpdate{IdempotencyKey: key, Values: values},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSelf(w, r, view)
}

func profileSelfKeysAllowed(values map[string]json.RawMessage) bool {
	if len(values) == 0 || len(values) > len(profileSelfWritableFields) {
		return false
	}
	for key := range values {
		if _, allowed := profileSelfWritableFields[key]; !allowed {
			return false
		}
	}
	return true
}
