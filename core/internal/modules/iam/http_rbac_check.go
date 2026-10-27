package iam

import "net/http"

func (handler *RBACHTTP) check(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if r.URL.RawQuery != "" {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[rbacCheckDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	input := RBACCheck{
		ResourceKey: document.ResourceSlug, ActionKey: document.ActionSlug,
	}
	if !rbacCheckValid(input) {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.Check(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, rbacEnvelope{Success: true, Data: view})
}
