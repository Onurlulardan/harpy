package platformops

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) downloadRecoveryKit(w http.ResponseWriter, r *http.Request) {
	operationID := r.PathValue("operation_id")
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil && !canonicalHTTPUUID(operationID) {
		err = errHTTPRequestInvalid
	}
	if err == nil {
		access, err = handler.accessFor(r, environment, "platform.recovery.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operation, err := handler.protection.GetControlOperation(r.Context(), operationID)
	if err != nil || operation.EnvironmentID != environment.EnvironmentID ||
		(operation.Kind != "enable" && operation.Kind != "rotate") {
		handler.writeFailure(w, r, recoveryOperationError(err))
		return
	}
	claim, err := handler.protection.ClaimKit(
		r.Context(), operationID, access.ActorID, access.SessionID)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeRecoveryKit(w, r, operationID, claim)
}

func (handler *HTTPHandler) ackRecoveryKit(w http.ResponseWriter, r *http.Request) {
	operationID := r.PathValue("operation_id")
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil && !canonicalHTTPUUID(operationID) {
		err = errHTTPRequestInvalid
	}
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment, "platform.recovery.read", "")
	}
	var operation platformopsapi.RecoveryControlOperation
	if err == nil {
		operation, err = handler.protection.GetControlOperation(r.Context(), operationID)
	}
	if err == nil && operation.EnvironmentID != environment.EnvironmentID {
		err = ErrHTTPForbidden
	}
	capability := "platform.recovery." + operation.Kind
	if err == nil && operation.Kind != "enable" && operation.Kind != "rotate" {
		err = platformopsapi.ErrConflict
	}
	if err == nil {
		access, err = handler.accessFor(r, environment, capability, "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	document, proof, err := decodeRecoveryKitAck(w, r)
	if err == nil {
		err = handler.protection.AcknowledgeKit(r.Context(), operationID,
			access.ActorID, access.SessionID, document.ArtifactSHA256, proof)
	}
	clear(proof)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	protection, err := handler.protection.GetProtection(
		r.Context(), environment.EnvironmentID)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, protection)
}

func recoveryOperationError(err error) error {
	if err != nil {
		return err
	}
	return platformopsapi.ErrConflict
}

func (handler *HTTPHandler) writeRecoveryKit(
	w http.ResponseWriter, r *http.Request, operationID string,
	claim platformopsapi.RecoveryKitClaim,
) {
	digest := sha256.Sum256(claim.ArtifactBytes)
	if !lowercaseSHA256(claim.ArtifactSHA256) ||
		fmt.Sprintf("%x", digest[:]) != claim.ArtifactSHA256 {
		handler.writeFailure(w, r, ErrHTTPUnavailable)
		return
	}
	header := w.Header()
	header.Set("Content-Type", "application/vnd.harpy.environment-recovery-kit.hpke.v1+json")
	header.Set("Content-Disposition", `attachment; filename="environment-recovery-kit-`+operationID+`.bin"`)
	header.Set("Digest", "sha-256=:"+base64.StdEncoding.EncodeToString(digest[:])+":")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Length", strconv.Itoa(len(claim.ArtifactBytes)))
	header.Set("X-Correlation-ID", platformOpsCorrelationID(r).String())
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(claim.ArtifactBytes); err != nil {
		handler.report(fmt.Errorf("write recovery kit: %w", err))
	}
}
