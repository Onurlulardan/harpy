package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

func (handler *HTTPHandler) requestRestore(w http.ResponseWriter, r *http.Request) {
	recoverySetID := r.PathValue("recovery_set_id")
	if !canonicalHTTPUUID(recoverySetID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery_sets.restore", recoverySetID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	document, err := decodeRestoreRequest(w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operationID := recoveryOperationID(uuid.MustParse(environment.EnvironmentID),
		uuid.MustParse(access.ActorID), "restore", key).String()
	operation, err := handler.restores.RequestPlannedRestore(r.Context(),
		platformopsapi.PlannedRestoreRequest{OperationID: operationID,
			RecoverySetID: recoverySetID, RequestedByTenantID: access.TenantID,
			RequestedByActorID:      access.ActorID,
			RequestedBySessionID:    access.SessionID,
			Disposition:             document.TargetDisposition,
			SafetyRecoverySetID:     nullableString(document.SafetyRecoverySetID),
			SafetyRecoverySetSHA256: nullableString(document.SafetyRecoverySetSHA256),
			WaiverReason:            nullableString(document.WaiverReason),
			ProbeSHA256:             nullableString(document.ProbeSHA256),
			DestructiveConfirmed:    *document.DestructiveConfirmed})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/platform/restore-operations/"+operation.OperationID)
	handler.writeSuccess(w, r, http.StatusAccepted, struct {
		OperationID string `json:"operation_id"`
		State       string `json:"state"`
	}{OperationID: operation.OperationID, State: "accepted"})
}

func (handler *HTTPHandler) getRestoreOperation(w http.ResponseWriter, r *http.Request) {
	operationID := r.PathValue("operation_id")
	if !canonicalHTTPUUID(operationID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.recovery_sets.read", "")
	}
	var operation platformopsapi.RestoreOperation
	if err == nil {
		operation, err = handler.restores.ResolveRestore(r.Context(), operationID)
	}
	if err == nil && operation.EnvironmentID != environment.EnvironmentID {
		err = platformopsapi.ErrNotFound
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, representRestore(operation))
}
