package iam

import "net/http"

type lifecyclePortalDecisionDocument struct {
	Decision string `json:"decision"`
}

func (handler *LifecycleHTTP) decidePortalSetup(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	bindingID, idValid := lifecyclePathUUID(r, "binding_id")
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	document, err := decodeIdentityDocument[lifecyclePortalDecisionDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if !idValid || !keyValid || !revisionValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	err = handler.application.DecidePortalSetup(r.Context(), claims,
		LifecyclePortalDecisionInput{LifecycleRevisionInput: LifecycleRevisionInput{
			ID: bindingID, ExpectedRevision: revision, IdempotencyKey: key},
			Decision: document.Decision})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
