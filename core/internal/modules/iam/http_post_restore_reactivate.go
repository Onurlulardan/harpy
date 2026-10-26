package iam

import (
	"encoding/json"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type postRestoreReactivateDocument struct {
	EnvironmentID      string          `json:"environment_id"`
	RestoreOperationID string          `json:"restore_operation_id"`
	OperationID        string          `json:"operation_id"`
	Credential         json.RawMessage `json:"credential"`
	NewPassword        json.RawMessage `json:"new_password"`
}

func (handler *PostRestoreHTTP) reactivateOwner(w http.ResponseWriter, r *http.Request) {
	document, err := decodeIdentityDocument[postRestoreReactivateDocument](w, r)
	if err != nil {
		handler.writePostRestoreInputFailure(w, r, err)
		return
	}
	defer clear(document.Credential)
	defer clear(document.NewPassword)
	environmentID, environmentValid := decodeUUID(document.EnvironmentID)
	restoreID, restoreValid := decodeUUID(document.RestoreOperationID)
	operationID, operationValid := decodeUUID(document.OperationID)
	origin, originValid := exactOrigin(r)
	password, passwordErr := decodeLoginPassword(document.NewPassword)
	if !environmentValid || !restoreValid || !operationValid || !originValid ||
		passwordErr != nil || !json.Valid(document.Credential) ||
		len(document.Credential) < 2 || document.Credential[0] != '{' {
		clear(password)
		handler.writePostRestoreInputFailure(w, r, httpinput.ErrInvalid)
		return
	}
	defer clear(password)
	result, err := handler.service.Reactivate(r.Context(), environmentID, restoreID,
		operationID, origin, document.Credential, password)
	if err != nil {
		handler.writePostRestoreApplicationFailure(w, r, err)
		return
	}
	grant, cookies, err := handler.issueRestoredOwnerSession(r, result, password, origin)
	if err != nil {
		handler.writePostRestoreApplicationFailure(w, r, err)
		return
	}
	appendSetCookies(w.Header(), cookies)
	handler.writePostRestoreJSON(w, r, http.StatusOK, postRestoreReactivateEnvelope{
		Success: true, Data: postRestoreReactivateData{UserID: result.UserID.String(),
			TenantID: result.TenantID.String(), SessionState: result.SessionState,
			Token: grant.Token, ExpiresAt: grant.ExpiresAt},
	})
}
