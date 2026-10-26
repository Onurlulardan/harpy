package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type postRestoreChallengeDocument struct {
	EnvironmentID      string `json:"environment_id"`
	RestoreOperationID string `json:"restore_operation_id"`
}

func (handler *PostRestoreHTTP) challenge(w http.ResponseWriter, r *http.Request) {
	document, err := decodeIdentityDocument[postRestoreChallengeDocument](w, r)
	if err != nil {
		handler.writePostRestoreInputFailure(w, r, err)
		return
	}
	environmentID, environmentValid := decodeUUID(document.EnvironmentID)
	restoreID, restoreValid := decodeUUID(document.RestoreOperationID)
	origin, originValid := exactOrigin(r)
	if !environmentValid || !restoreValid || !originValid {
		handler.writePostRestoreInputFailure(w, r, httpinput.ErrInvalid)
		return
	}
	result, err := handler.service.Challenge(r.Context(), environmentID, restoreID, origin)
	if err != nil {
		handler.writePostRestoreApplicationFailure(w, r, err)
		return
	}
	handler.writePostRestoreJSON(w, r, http.StatusOK, postRestoreChallengeEnvelope{
		Success: true, Data: postRestoreChallengeData{OperationID: result.OperationID.String(),
			PublicKey: result.PublicKey, ExpiresAt: result.ExpiresAt.UTC()},
	})
}
