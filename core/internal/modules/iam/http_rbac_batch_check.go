package iam

import "net/http"

func (handler *RBACHTTP) batchCheck(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if r.URL.RawQuery != "" {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[rbacBatchCheckDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	input := RBACBatchCheck{Checks: make([]RBACCheck, len(document.Checks))}
	for index, check := range document.Checks {
		input.Checks[index] = RBACCheck{
			ResourceKey: check.ResourceSlug, ActionKey: check.ActionSlug,
		}
	}
	if !rbacBatchCheckValid(input) {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.BatchCheck(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, rbacEnvelope{Success: true, Data: view})
}
