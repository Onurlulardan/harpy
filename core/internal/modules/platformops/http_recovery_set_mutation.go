package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

func (handler *HTTPHandler) createRecoverySet(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery_sets.create", environment.EnvironmentID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	environmentID := uuid.MustParse(environment.EnvironmentID)
	actorID := uuid.MustParse(access.ActorID)
	operationID := recoveryOperationID(environmentID, actorID, "recovery-set-create", key)
	recoverySetID := recoveryOperationID(environmentID, actorID, "recovery-set", key)
	set, err := handler.recoverySets.RequestRecoverySet(r.Context(), platformopsapi.RecoverySetRequest{
		RecoverySetID: recoverySetID.String(), OperationID: operationID.String(),
		EnvironmentID: environment.EnvironmentID})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/platform/recovery-sets/"+set.RecoverySetID)
	handler.writeSuccess(w, r, http.StatusAccepted, struct {
		RecoverySetID string `json:"recovery_set_id"`
		OperationID   string `json:"operation_id"`
		EnvironmentID string `json:"environment_id"`
		Sequence      int64  `json:"sequence"`
		State         string `json:"state"`
	}{RecoverySetID: set.RecoverySetID, OperationID: set.OperationID,
		EnvironmentID: set.EnvironmentID, Sequence: set.Sequence, State: set.State})
}

func (handler *HTTPHandler) deleteRecoverySet(w http.ResponseWriter, r *http.Request) {
	recoverySetID := r.PathValue("recovery_set_id")
	if !canonicalHTTPUUID(recoverySetID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery_sets.delete", recoverySetID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	set, err := handler.recoverySets.GetRecoverySet(r.Context(), recoverySetID)
	if err == nil && set.EnvironmentID != environment.EnvironmentID {
		err = platformopsapi.ErrNotFound
	}
	if err == nil && set.State == "protected" {
		err = handler.recoverySets.MarkRecoverySetDeleting(
			r.Context(), set.RecoverySetID, set.ManifestSHA256)
	} else if err == nil && set.State != "deleting" && set.State != "deleted" {
		err = platformopsapi.ErrConflict
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operationID := recoveryOperationID(uuid.MustParse(environment.EnvironmentID),
		uuid.MustParse(access.ActorID), "recovery-set-delete", key).String()
	w.Header().Set("Location", "/api/platform/recovery-sets/"+set.RecoverySetID)
	handler.writeSuccess(w, r, http.StatusAccepted, struct {
		OperationID string `json:"operation_id"`
		State       string `json:"state"`
	}{OperationID: operationID, State: "accepted"})
}
