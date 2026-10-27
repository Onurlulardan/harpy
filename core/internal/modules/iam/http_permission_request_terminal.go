package iam

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

type permissionRequestTerminalCall func(
	context.Context,
	accesstoken.Claims,
	PermissionRequestTerminal,
) (PermissionRequestRefView, error)

func (handler *PermissionRequestHTTP) cancel(w http.ResponseWriter, r *http.Request) {
	handler.handleTerminal(w, r, handler.application.Cancel)
}

func (handler *PermissionRequestHTTP) expire(w http.ResponseWriter, r *http.Request) {
	handler.handleTerminal(w, r, handler.application.Expire)
}

func (handler *PermissionRequestHTTP) handleTerminal(
	w http.ResponseWriter,
	r *http.Request,
	execute permissionRequestTerminalCall,
) {
	claims, authenticated := handler.claims(r)
	id, idValid := identityPathID(r)
	revision, revisionValid := identityExpectedRevision(r)
	if !authenticated {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if execute == nil || !idValid || !revisionValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[permissionRequestStateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	view, err := execute(r.Context(), claims, PermissionRequestTerminal{
		ID: id, ExpectedRevision: revision,
		ExpectedState: permissionrequest.State(document.ExpectedState),
	})
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
