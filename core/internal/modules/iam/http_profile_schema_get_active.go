package iam

import "net/http"

func (handler *ProfileHTTP) getActiveSchema(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	resource, valid := profilePathResource(r)
	if !valid {
		handler.writeFailure(w, r, ErrProfileHTTPInvalid)
		return
	}
	view, err := handler.application.GetActiveSchema(r.Context(), claims, resource)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSchema(w, r, view)
}
