package iam

import "net/http"

func (handler *LifecycleHTTP) issueRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	key, keyValid := identityIdempotencyKey(r)
	document, err := decodeIdentityDocument[lifecycleEnrollmentStartDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if !keyValid {
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	view, err := handler.application.IssueRecoveryCodes(r.Context(), claims,
		LifecycleRecoveryIssueInput{IdempotencyKey: key, DisplayLabel: document.DisplayLabel})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusCreated, view)
}
