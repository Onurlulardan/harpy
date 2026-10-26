package iam

import "net/http"

func (handler *RBACHTTP) deleteResource(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	mutation, valid := rbacRevisionedMutation(r)
	if !valid {
		handler.writeFailure(w, r, identityFailurePrecondition)
		return
	}
	if err := handler.application.DeleteResource(
		r.Context(), claims, mutation,
	); err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeDelete(w, r)
}
