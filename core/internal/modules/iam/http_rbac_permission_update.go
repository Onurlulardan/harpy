package iam

import "net/http"

func (handler *RBACHTTP) updatePermission(w http.ResponseWriter, r *http.Request) {
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
	document, err := decodeIdentityDocument[rbacPermissionUpdateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	current, err := handler.application.GetPermission(r.Context(), claims, mutation.ID)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	input, valid := rbacMergePermissionUpdate(document, current, mutation)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.UpdatePermission(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
