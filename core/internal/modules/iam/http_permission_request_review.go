package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (handler *PermissionRequestHTTP) review(w http.ResponseWriter, r *http.Request) {
	claims, authenticated := handler.claims(r)
	id, idValid := identityPathID(r)
	revision, revisionValid := identityExpectedRevision(r)
	key, keyValid := identityIdempotencyKey(r)
	if !authenticated {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !idValid || !revisionValid || !keyValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[permissionRequestStateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	view, err := handler.application.Review(r.Context(), claims, PermissionRequestReview{
		ID: id, ExpectedRevision: revision,
		ExpectedState:  permissionrequest.State(document.ExpectedState),
		IdempotencyKey: key,
	})
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
