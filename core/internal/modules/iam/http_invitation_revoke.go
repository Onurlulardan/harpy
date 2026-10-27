package iam

import "net/http"

func (handler *InvitationHTTP) revoke(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errInvitationAuthentication)
		return
	}
	id, idValid := identityPathID(r)
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	if !idValid || !keyValid || !revisionValid {
		handler.writeFailure(w, r, ErrInvitationHTTPInvalid)
		return
	}
	err := handler.application.Revoke(
		r.Context(), claims, InvitationRevokeInput{
			ID: id, ExpectedRevision: revision, IdempotencyKey: key,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, invitationMessage{
		Message: "Invitation cancelled",
	})
}
