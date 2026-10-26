package platformops

import (
	"encoding/json"
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type recoveryAnchorRegistrationDocument struct {
	OperationID string          `json:"operation_id"`
	Credential  json.RawMessage `json:"credential"`
}

func (handler *HTTPHandler) registerRecoveryAnchor(
	w http.ResponseWriter,
	r *http.Request,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment,
			"platform.recovery.anchor.rotate", environment.EnvironmentID)
	}
	key, valid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !valid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	var document recoveryAnchorRegistrationDocument
	if err = httpinput.DecodeJSON(w, r, 64*1024, &document); err != nil ||
		!canonicalHTTPUUID(document.OperationID) || len(document.Credential) < 2 ||
		document.Credential[0] != '{' {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	digest, err := recoveryRequestSHA256(document)
	if err != nil {
		handler.writeFailure(w, r, ErrHTTPUnavailable)
		return
	}
	operation, err := handler.anchors.RegisterRecoveryAnchor(
		r.Context(), platformopsapi.RecoveryAnchorRegisterCommand{
			RecoveryAnchorSubject: platformopsapi.RecoveryAnchorSubject{
				EnvironmentID: environment.EnvironmentID,
				TenantID:      environment.TenantID, ActorID: access.ActorID},
			OperationID: document.OperationID, IdempotencyKey: key,
			RequestSHA256:  digest,
			CredentialJSON: append([]byte(nil), document.Credential...)})
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
