package iam

import "net/http"

func (handler *ProfileHTTP) publishSchema(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	resource, resourceValid := profilePathResource(r)
	key, keyValid := identityIdempotencyKey(r)
	if !resourceValid || !keyValid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	document, err := decodeIdentityDocument[profileSchemaPublishDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	input, valid := profilePublishInput(resource, key, document)
	if !valid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	view, err := handler.application.PublishSchema(r.Context(), claims, input)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeMutation(w, r, http.StatusCreated, view)
}
