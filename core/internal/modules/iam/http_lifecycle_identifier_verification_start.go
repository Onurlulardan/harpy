package iam

import "net/http"

func (handler *LifecycleHTTP) startEmailVerification(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	id, idValid := lifecyclePathUUID(r, "id")
	key, keyValid := identityIdempotencyKey(r)
	if !idValid || !keyValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	if err := handler.application.StartEmailVerification(r.Context(), claims,
		LifecycleResourceInput{ID: id, IdempotencyKey: key}); err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
