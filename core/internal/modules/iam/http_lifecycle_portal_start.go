package iam

import "net/http"

func (handler *LifecycleHTTP) startPortalSetup(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	bindingID, idValid := lifecyclePathUUID(r, "binding_id")
	key, keyValid := identityIdempotencyKey(r)
	if !idValid || !keyValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	if err := handler.application.StartPortalSetup(r.Context(), claims,
		LifecycleResourceInput{ID: bindingID, IdempotencyKey: key}); err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
