package iam

import "net/http"

func (handler *ProfileHTTP) getSchema(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	resource, resourceValid := profilePathResource(r)
	schemaID, schemaValid := profilePathUUID(r, "schema_id")
	if !resourceValid || !schemaValid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	view, err := handler.application.GetSchema(
		r.Context(), claims, resource, schemaID,
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSchema(w, r, view)
}
