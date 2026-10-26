package iam

import (
	"fmt"
	"net/http"
)

func (handler *LifecycleHTTP) readPortalSetup(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	bindingID, valid := lifecyclePathUUID(r, "binding_id")
	if !valid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	setup, err := handler.application.ReadPortalSetup(r.Context(), claims, bindingID)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	view, valid := lifecyclePortalSetupResponse(setup)
	if !valid || view.BindingID != bindingID {
		handler.writeFailure(w, r, ErrLifecycleHTTPUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeJSON(w, r, http.StatusOK, view)
}
