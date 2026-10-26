package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) listRecoveryAnchors(
	w http.ResponseWriter,
	r *http.Request,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery.anchor.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	anchors, err := handler.anchors.ListRecoveryAnchors(
		r.Context(), platformopsapi.RecoveryAnchorSubject{
			EnvironmentID: environment.EnvironmentID,
			TenantID:      environment.TenantID, ActorID: access.ActorID})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, representRecoveryAnchors(anchors))
}
