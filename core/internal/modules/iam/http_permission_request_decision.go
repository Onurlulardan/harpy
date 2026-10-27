package iam

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

type permissionRequestDecisionCall func(
	context.Context,
	accesstoken.Claims,
	PermissionRequestDecision,
) (PermissionRequestRefView, error)

func (handler *PermissionRequestHTTP) approve(w http.ResponseWriter, r *http.Request) {
	handler.handleDecision(w, r, handler.application.Approve)
}

func (handler *PermissionRequestHTTP) reject(w http.ResponseWriter, r *http.Request) {
	handler.handleDecision(w, r, handler.application.Reject)
}

func (handler *PermissionRequestHTTP) handleDecision(
	w http.ResponseWriter,
	r *http.Request,
	execute permissionRequestDecisionCall,
) {
	claims, authenticated := handler.claims(r)
	id, idValid := identityPathID(r)
	revision, revisionValid := identityExpectedRevision(r)
	key, keyValid := identityIdempotencyKey(r)
	if !authenticated {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if execute == nil || !idValid || !revisionValid || !keyValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[permissionRequestDecisionDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	taskID, valid := decodeUUID(document.ReviewTaskID)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := execute(r.Context(), claims, PermissionRequestDecision{
		ID: id, ExpectedRevision: revision,
		ExpectedState: permissionrequest.State(document.ExpectedState),
		ReviewTaskID:  taskID, IdempotencyKey: key,
	})
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
