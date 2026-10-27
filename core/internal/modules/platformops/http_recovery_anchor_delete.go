package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

func (handler *HTTPHandler) deleteRecoveryAnchor(
	w http.ResponseWriter,
	r *http.Request,
) {
	anchorID := r.PathValue("anchor_id")
	if !canonicalHTTPUUID(anchorID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery.anchor.rotate", anchorID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	digest, err := recoveryRequestSHA256(struct {
		AnchorID string `json:"anchor_id"`
	}{AnchorID: anchorID})
	if err != nil {
		handler.writeFailure(w, r, ErrHTTPUnavailable)
		return
	}
	operationID := recoveryOperationID(uuid.MustParse(environment.EnvironmentID),
		uuid.MustParse(access.ActorID), "anchor-delete:"+anchorID, key).String()
	operation, err := handler.anchors.DeleteRecoveryAnchor(
		r.Context(), platformopsapi.RecoveryAnchorDeleteCommand{
			RecoveryAnchorSubject: platformopsapi.RecoveryAnchorSubject{
				EnvironmentID: environment.EnvironmentID,
				TenantID:      environment.TenantID, ActorID: access.ActorID},
			OperationID: operationID, AnchorID: anchorID,
			IdempotencyKey: key, RequestSHA256: digest})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/platform/recovery-operations/"+operation.OperationID)
	handler.writeSuccess(w, r, http.StatusAccepted, struct {
		OperationID string `json:"operation_id"`
		State       string `json:"state"`
	}{OperationID: operation.OperationID, State: operation.State})
}
