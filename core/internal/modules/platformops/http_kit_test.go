package platformops

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryKitDownloadIsSessionBoundCanonicalArtifact(t *testing.T) {
	handler, access, protection := httpHandlerFixture()
	artifact := []byte(`{"schema":"environment-recovery-kit.hpke.v1"}`)
	digest := sha256.Sum256(artifact)
	protection.operation.State, protection.operation.KitClaimState = "kit_ready", "ready"
	protection.operation.ArtifactSHA256 = fmt.Sprintf("%x", digest[:])
	protection.claim.ArtifactBytes = artifact
	protection.claim.ArtifactSHA256 = protection.operation.ArtifactSHA256
	request := httptest.NewRequest(http.MethodPost,
		"/api/platform/recovery-operations/"+httpOperationID+"/kit-download", nil)
	request.SetPathValue("operation_id", httpOperationID)
	request.Header.Set("Authorization", "Bearer fixture")
	recorder := httptest.NewRecorder()
	handler.downloadRecoveryKit(recorder, request)
	if recorder.Code != http.StatusOK ||
		recorder.Header().Get("Content-Type") !=
			"application/vnd.harpy.environment-recovery-kit.hpke.v1+json" ||
		recorder.Header().Get("Digest") !=
			"sha-256=:"+base64.StdEncoding.EncodeToString(digest[:])+":" ||
		!bytesEqual(recorder.Body.Bytes(), artifact) {
		t.Fatalf("kit response=%d %s", recorder.Code, recorder.Body.String())
	}
	if len(access.requests) != 1 || access.requests[0].Capability != "platform.recovery.read" ||
		access.requests[0].RequireStepUp {
		t.Fatal("kit claim access contract drifted")
	}
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
