package iam

import "net/http"

func (handler *LifecycleHTTP) listRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	items, err := handler.application.ListRecoveryCodes(r.Context(), claims)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, map[string]any{"items": items})
}
