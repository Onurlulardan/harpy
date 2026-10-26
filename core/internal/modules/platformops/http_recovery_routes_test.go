package platformops

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type recoveryRoutesFixture struct {
	httpRecoveryFixture
	setRequest     platformopsapi.RecoverySetRequest
	restoreRequest platformopsapi.PlannedRestoreRequest
	artifactClaim  platformopsapi.RecoveryArtifactClaim
	artifact       platformopsapi.RecoveryArtifact
}

func (fixture *recoveryRoutesFixture) RequestRecoverySet(
	_ context.Context, request platformopsapi.RecoverySetRequest,
) (platformopsapi.RecoverySet, error) {
	fixture.setRequest = request
	return platformopsapi.RecoverySet{RecoverySetID: request.RecoverySetID,
		OperationID: request.OperationID, EnvironmentID: request.EnvironmentID,
		Sequence: 1, State: "requested"}, nil
}

func (fixture *recoveryRoutesFixture) GetRecoverySet(
	context.Context, string,
) (platformopsapi.RecoverySet, error) {
	protectedAt := time.Date(2026, 7, 23, 6, 0, 0, 0, time.UTC)
	return platformopsapi.RecoverySet{RecoverySetID: httpOperationID,
		OperationID: httpOperationID, EnvironmentID: httpEnvironmentID,
		Sequence: 1, State: "protected", ManifestSHA256: fixture.artifact.SHA256,
		ProtectedAt: &protectedAt}, nil
}

func (fixture *recoveryRoutesFixture) ClaimRecoveryManifest(
	_ context.Context, claim platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	fixture.artifactClaim = claim
	return fixture.artifact, nil
}

func (fixture *recoveryRoutesFixture) RequestPlannedRestore(
	_ context.Context, request platformopsapi.PlannedRestoreRequest,
) (platformopsapi.RestoreOperation, error) {
	fixture.restoreRequest = request
	return platformopsapi.RestoreOperation{OperationID: request.OperationID,
		RecoverySetID: request.RecoverySetID, EnvironmentID: httpEnvironmentID,
		Disposition: request.Disposition, State: "plan_ready"}, nil
}

func (fixture *recoveryRoutesFixture) ClaimRestorePlan(
	_ context.Context, claim platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	fixture.artifactClaim = claim
	return fixture.artifact, nil
}

func newRecoveryRoutesHandler(
	t *testing.T, fixture *recoveryRoutesFixture,
) (*HTTPHandler, *httpAccessFixture) {
	t.Helper()
	access := &httpAccessFixture{value: HTTPAccess{ActorID: httpActorID,
		TenantID: httpTenantID, SessionID: httpSessionID,
		TokenVersion: 1, AuthorizationVersion: 1, PlatformClock: 1}}
	protection := &httpProtectionFixture{}
	operations := &httpRecoveryFixture{}
	handler, err := NewHTTPHandler(HTTPDependencies{Access: access,
		Environment: httpEnvironmentFixture{value: HTTPEnvironment{
			EnvironmentID: httpEnvironmentID, TenantID: httpTenantID,
			DeploymentMode: "onpremise"}}, Diagnostics: fixture,
		Protection: protection, RecoverySets: fixture, Restores: fixture,
		Anchors:    &httpRecoveryAnchorFixture{},
		Operations: operations, Traces: operations, Alerts: operations,
		PermissionCache: operations,
		Report:          func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	return handler, access
}

func TestCreateRecoverySetUsesStableIdentityAndExactStepUp(t *testing.T) {
	fixture := &recoveryRoutesFixture{}
	handler, access := newRecoveryRoutesHandler(t, fixture)
	request := httptest.NewRequest(http.MethodPost, "/api/platform/recovery-sets", nil)
	request.Header.Set("Idempotency-Key", "recovery-create-0001")
	response := httptest.NewRecorder()
	handler.createRecoverySet(response, request)
	if response.Code != http.StatusAccepted || fixture.setRequest.EnvironmentID != httpEnvironmentID {
		t.Fatalf("create recovery set = %d / %#v", response.Code, fixture.setRequest)
	}
	first := fixture.setRequest
	handler.createRecoverySet(httptest.NewRecorder(), request.Clone(request.Context()))
	if fixture.setRequest != first || len(access.requests) != 2 {
		t.Fatal("recovery-set idempotency was not stable")
	}
	want, _ := platformopsapi.StepUpPurpose(httpTenantID,
		"platform.recovery_sets.create", httpEnvironmentID)
	if !access.requests[0].RequireStepUp || access.requests[0].StepUpPurpose != want {
		t.Fatal("recovery-set create step-up contract drifted")
	}
}

func TestRecoveryArtifactDownloadsAreSessionBound(t *testing.T) {
	document := []byte(`{"schema":"artifact.v1"}`)
	digest := sha256.Sum256(document)
	fixture := &recoveryRoutesFixture{artifact: platformopsapi.RecoveryArtifact{
		SHA256: fmt.Sprintf("%x", digest[:]), Bytes: document}}
	handler, access := newRecoveryRoutesHandler(t, fixture)
	request := httptest.NewRequest(http.MethodPost, "/manifest-download", nil)
	request.SetPathValue("recovery_set_id", httpOperationID)
	response := httptest.NewRecorder()
	handler.downloadRecoveryManifest(response, request)
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), document) ||
		fixture.artifactClaim.ActorID != httpActorID ||
		fixture.artifactClaim.SessionID != httpSessionID {
		t.Fatalf("manifest download = %d / %#v", response.Code, fixture.artifactClaim)
	}
	if access.requests[0].Capability != "platform.recovery_sets.manifest_download" ||
		!access.requests[0].RequireStepUp || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("manifest download access or response contract drifted")
	}
}

func TestRestoreRequestBindsOwnerSessionAndExactEvidence(t *testing.T) {
	fixture := &recoveryRoutesFixture{}
	handler, _ := newRecoveryRoutesHandler(t, fixture)
	body := `{"target_disposition":"existing_readable",` +
		`"safety_recovery_set_id":"018f0000-0000-7000-8000-000000000006",` +
		`"safety_recovery_set_sha256":"` +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		`","waiver_reason":null,"probe_sha256":null,"destructive_confirmed":false}`
	request := httptest.NewRequest(http.MethodPost, "/restore", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "restore-request-0001")
	request.SetPathValue("recovery_set_id", httpOperationID)
	response := httptest.NewRecorder()
	handler.requestRestore(response, request)
	if response.Code != http.StatusAccepted || fixture.restoreRequest.RequestedByActorID != httpActorID ||
		fixture.restoreRequest.RequestedByTenantID != httpTenantID ||
		fixture.restoreRequest.RequestedBySessionID != httpSessionID ||
		fixture.restoreRequest.SafetyRecoverySetID == "" || fixture.restoreRequest.DestructiveConfirmed {
		t.Fatalf("restore request = %d / %#v", response.Code, fixture.restoreRequest)
	}
}
