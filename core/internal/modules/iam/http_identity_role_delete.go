package iam

import "net/http"

func (handler *IdentityHTTP) deleteRole(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	mutation, valid := identityMutationFromRequest(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailurePrecondition)
		return
	}
	if err := handler.application.DeleteRole(r.Context(), claims, mutation); err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	handler.writeDelete(w, r, 1)
}
