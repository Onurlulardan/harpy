package iam

import "net/http"

func (handler *LifecycleHTTP) listAuthenticators(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	page, pageValid := lifecyclePage(r)
	if !pageValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	items, err := handler.application.ListAuthenticators(r.Context(), claims, page)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, map[string]any{"items": items})
}
