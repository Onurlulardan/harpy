package iam

import "net/http"

func (handler *ProfileHTTP) retireField(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	resource, resourceValid := profilePathResource(r)
	fieldID, fieldValid := profilePathUUID(r, "field_id")
	key, keyValid := identityIdempotencyKey(r)
	if !resourceValid || !fieldValid || !keyValid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	document, err := decodeIdentityDocument[profileSchemaRetireDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	schemaID, schemaValid := decodeUUID(document.SchemaID)
	activeID, activeValid := decodeUUID(document.ExpectedActiveSchemaID)
	if !schemaValid || !activeValid || document.ExpectedFieldRevision < 1 {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	view, err := handler.application.RetireField(
		r.Context(), claims, ProfileSchemaRetire{
			IdempotencyKey: key, SchemaID: schemaID, Resource: resource,
			ExpectedActiveSchemaID: activeID, FieldID: fieldID,
			ExpectedFieldRevision: document.ExpectedFieldRevision,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeMutation(w, r, http.StatusOK, view)
}
