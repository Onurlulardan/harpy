package iam

import "net/http"

func (handler *RBACHTTP) listPermissions(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	input, valid := rbacListRequest(r, true)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	page, err := handler.application.ListPermissions(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writePage(w, r, page.Items, page.HasMore, page.NextAfterID)
}
