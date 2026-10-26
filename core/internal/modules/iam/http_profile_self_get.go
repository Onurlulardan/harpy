package iam

import "net/http"

func (handler *ProfileHTTP) getSelf(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errProfileAuthentication)
		return
	}
	view, err := handler.application.GetSelf(r.Context(), claims)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSelf(w, r, view)
}
