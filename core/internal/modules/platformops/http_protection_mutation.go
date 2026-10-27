package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

func (handler *HTTPHandler) enableProtection(w http.ResponseWriter, r *http.Request) {
	handler.requestProtection(w, r, "enable", "platform.recovery.enable")
}

func (handler *HTTPHandler) rotateProtection(w http.ResponseWriter, r *http.Request) {
	handler.requestProtection(w, r, "rotate", "platform.recovery.rotate")
}

func (handler *HTTPHandler) disableProtection(w http.ResponseWriter, r *http.Request) {
	handler.requestProtection(w, r, "disable", "platform.recovery.disable")
}

func (handler *HTTPHandler) requestProtection(
	w http.ResponseWriter, r *http.Request, kind, capability string,
) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(
			r, environment, capability, environment.EnvironmentID)
	}
	key, keyValid := exactIdempotencyKey(r.Header.Values("Idempotency-Key"))
	if err != nil || !keyValid {
		handler.writeFailure(w, r, firstPlatformOpsError(err, httpinput.ErrInvalid))
		return
	}
	request, err := handler.decodeRecoveryControl(w, r, environment, access, kind, key)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operation, err := handler.protection.RequestControl(r.Context(), request)
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

func (handler *HTTPHandler) decodeRecoveryControl(
	w http.ResponseWriter, r *http.Request, environment HTTPEnvironment,
	access HTTPAccess, kind, key string,
) (platformopsapi.RecoveryControlRequest, error) {
	environmentID := uuid.MustParse(environment.EnvironmentID)
	actorID := uuid.MustParse(access.ActorID)
	request := platformopsapi.RecoveryControlRequest{
		OperationID:   recoveryOperationID(environmentID, actorID, kind, key).String(),
		EnvironmentID: environment.EnvironmentID, ActorUserID: access.ActorID,
		IdempotencyKey: key, Kind: kind,
	}
	var digestInput any = struct{}{}
	if kind != "disable" {
		document, publicKey, err := decodeRecoveryClaim(w, r)
		if err != nil {
			return platformopsapi.RecoveryControlRequest{}, err
		}
		request.ClaimAlgorithm = document.Claim.Algorithm
		request.ClaimPublicKey = publicKey
		digestInput = document
	}
	digest, err := recoveryRequestSHA256(digestInput)
	if err != nil {
		clear(request.ClaimPublicKey)
		return platformopsapi.RecoveryControlRequest{}, ErrHTTPUnavailable
	}
	request.RequestSHA256 = digest
	return request, nil
}

func firstPlatformOpsError(actual, fallback error) error {
	if actual != nil {
		return actual
	}
	return fallback
}
