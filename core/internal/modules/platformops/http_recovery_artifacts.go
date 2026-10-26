package platformops

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) downloadRecoveryManifest(w http.ResponseWriter, r *http.Request) {
	handler.downloadRecoveryArtifact(w, r, r.PathValue("recovery_set_id"),
		"platform.recovery_sets.manifest_download", "recovery-manifest",
		"application/vnd.harpy.recovery-manifest.v1+json", true)
}

func (handler *HTTPHandler) downloadRestorePlan(w http.ResponseWriter, r *http.Request) {
	handler.downloadRecoveryArtifact(w, r, r.PathValue("operation_id"),
		"platform.restore_operations.plan_download", "restore-plan",
		"application/vnd.harpy.restore-plan.v1+json", false)
}

func (handler *HTTPHandler) downloadRecoveryArtifact(
	w http.ResponseWriter, r *http.Request, artifactID, capability, filename, mediaType string,
	manifest bool,
) {
	if !canonicalHTTPUUID(artifactID) {
		handler.writeFailure(w, r, platformopsapi.ErrInvalid)
		return
	}
	environment, err := handler.onPremiseEnvironmentFor(r)
	var access HTTPAccess
	if err == nil {
		access, err = handler.accessFor(r, environment, capability, artifactID)
	}
	var artifact platformopsapi.RecoveryArtifact
	claim := platformopsapi.RecoveryArtifactClaim{ArtifactID: artifactID,
		TenantID: environment.TenantID, ActorID: access.ActorID, SessionID: access.SessionID}
	if err == nil && manifest {
		artifact, err = handler.recoverySets.ClaimRecoveryManifest(r.Context(), claim)
	} else if err == nil {
		artifact, err = handler.restores.ClaimRestorePlan(r.Context(), claim)
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeRecoveryArtifact(w, r, artifactID, filename, mediaType, artifact)
}

func (handler *HTTPHandler) writeRecoveryArtifact(
	w http.ResponseWriter, r *http.Request, artifactID, filename, mediaType string,
	artifact platformopsapi.RecoveryArtifact,
) {
	digest := sha256.Sum256(artifact.Bytes)
	if !lowercaseSHA256(artifact.SHA256) || fmt.Sprintf("%x", digest[:]) != artifact.SHA256 {
		handler.writeFailure(w, r, ErrHTTPUnavailable)
		return
	}
	header := w.Header()
	header.Set("Content-Type", mediaType)
	header.Set("Content-Disposition", `attachment; filename="`+filename+`-`+artifactID+`.json"`)
	header.Set("Digest", "sha-256=:"+base64.StdEncoding.EncodeToString(digest[:])+":")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Length", strconv.Itoa(len(artifact.Bytes)))
	header.Set("X-Correlation-ID", platformOpsCorrelationID(r).String())
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(artifact.Bytes); err != nil {
		handler.report(fmt.Errorf("write recovery artifact: %w", err))
	}
}
