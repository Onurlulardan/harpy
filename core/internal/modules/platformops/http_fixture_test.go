package platformops

import (
	"context"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

const (
	httpEnvironmentID = "018f0000-0000-7000-8000-000000000001"
	httpTenantID      = "018f0000-0000-7000-8000-000000000002"
	httpActorID       = "018f0000-0000-7000-8000-000000000003"
	httpSessionID     = "018f0000-0000-7000-8000-000000000004"
	httpOperationID   = "018f0000-0000-7000-8000-000000000005"
)

type httpEnvironmentFixture struct{ value HTTPEnvironment }

func (fixture httpEnvironmentFixture) Current(context.Context) (HTTPEnvironment, error) {
	return fixture.value, nil
}

type httpAccessFixture struct {
	requests []HTTPAccessRequest
	value    HTTPAccess
}

func (fixture *httpAccessFixture) Verify(
	_ context.Context, request HTTPAccessRequest,
) (HTTPAccess, error) {
	fixture.requests = append(fixture.requests, request)
	return fixture.value, nil
}

type httpProtectionFixture struct {
	request   platformopsapi.RecoveryControlRequest
	operation platformopsapi.RecoveryControlOperation
	claim     platformopsapi.RecoveryKitClaim
	ack       struct {
		operationID, actorID, sessionID, digest string
		proof                                   []byte
	}
}

type httpRecoveryFixture struct{}

func (*httpRecoveryFixture) ReadDiagnostic(
	context.Context, platformopsapi.DiagnosticRequest,
) (platformopsapi.DiagnosticSnapshot, error) {
	return platformopsapi.DiagnosticSnapshot{ID: platformopsapi.DiagnosticDatabaseStats,
		OutputSchemaSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ObservedAt:         time.Date(2026, 7, 23, 5, 0, 0, 0, time.UTC)}, nil
}

func (*httpRecoveryFixture) FlushPermissionCache(
	context.Context, platformopsapi.PermissionCacheFlush,
) (platformopsapi.PermissionCacheResult, error) {
	return platformopsapi.PermissionCacheResult{
		AuthorizationClock: 2, Revision: 2}, nil
}

func (*httpRecoveryFixture) RequestRecoverySet(
	context.Context, platformopsapi.RecoverySetRequest,
) (platformopsapi.RecoverySet, error) {
	return platformopsapi.RecoverySet{}, nil
}

func (*httpRecoveryFixture) GetRecoverySet(
	context.Context, string,
) (platformopsapi.RecoverySet, error) {
	return platformopsapi.RecoverySet{}, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) ListRecoverySets(
	context.Context, string,
) ([]platformopsapi.RecoverySet, error) {
	return []platformopsapi.RecoverySet{}, nil
}

func (*httpRecoveryFixture) ClaimRecoveryManifest(
	context.Context, platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	return platformopsapi.RecoveryArtifact{}, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) MarkRecoverySetDeleting(context.Context, string, string) error {
	return nil
}

func (*httpRecoveryFixture) PruneRecoverySets(
	context.Context, string,
) (platformopsapi.RetentionResult, error) {
	return platformopsapi.RetentionResult{}, nil
}

func (*httpRecoveryFixture) RequestRestore(
	context.Context, platformopsapi.RestoreOperationRequest,
) (platformopsapi.RestoreOperation, error) {
	return platformopsapi.RestoreOperation{}, nil
}

func (*httpRecoveryFixture) RequestPlannedRestore(
	context.Context, platformopsapi.PlannedRestoreRequest,
) (platformopsapi.RestoreOperation, error) {
	return platformopsapi.RestoreOperation{}, nil
}

func (*httpRecoveryFixture) ResolveRestore(
	context.Context, string,
) (platformopsapi.RestoreOperation, error) {
	return platformopsapi.RestoreOperation{}, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) ClaimRestorePlan(
	context.Context, platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	return platformopsapi.RecoveryArtifact{}, platformopsapi.ErrNotFound
}

func (fixture *httpProtectionFixture) GetProtection(
	context.Context, string,
) (platformopsapi.RecoveryProtection, error) {
	return platformopsapi.RecoveryProtection{EnvironmentID: httpEnvironmentID,
		State: "protected", Generation: 1, KitAcknowledged: true,
		UpdatedAt: time.Date(2026, 7, 23, 5, 0, 0, 0, time.UTC)}, nil
}

func (fixture *httpProtectionFixture) RequestControl(
	_ context.Context, request platformopsapi.RecoveryControlRequest,
) (platformopsapi.RecoveryControlOperation, error) {
	fixture.request = request
	return fixture.operation, nil
}

func (fixture *httpProtectionFixture) GetControlOperation(
	context.Context, string,
) (platformopsapi.RecoveryControlOperation, error) {
	return fixture.operation, nil
}

func (fixture *httpProtectionFixture) ClaimKit(
	context.Context, string, string, string,
) (platformopsapi.RecoveryKitClaim, error) {
	return fixture.claim, nil
}

func (fixture *httpProtectionFixture) AcknowledgeKit(
	_ context.Context, operationID, actorID, sessionID, digest string, proof []byte,
) error {
	fixture.ack.operationID, fixture.ack.actorID = operationID, actorID
	fixture.ack.sessionID, fixture.ack.digest = sessionID, digest
	fixture.ack.proof = append([]byte(nil), proof...)
	return nil
}

func httpHandlerFixture() (*HTTPHandler, *httpAccessFixture, *httpProtectionFixture) {
	access := &httpAccessFixture{value: HTTPAccess{ActorID: httpActorID,
		TenantID: httpTenantID, SessionID: httpSessionID,
		TokenVersion: 1, AuthorizationVersion: 1, PlatformClock: 1}}
	protection := &httpProtectionFixture{operation: platformopsapi.RecoveryControlOperation{
		OperationID: httpOperationID, EnvironmentID: httpEnvironmentID,
		Kind: "enable", State: "accepted", KitClaimState: "pending",
		CreatedAt: time.Date(2026, 7, 23, 5, 0, 0, 0, time.UTC)}}
	recovery := &httpRecoveryFixture{}
	handler, _ := NewHTTPHandler(HTTPDependencies{Access: access,
		Environment: httpEnvironmentFixture{value: HTTPEnvironment{
			EnvironmentID: httpEnvironmentID, TenantID: httpTenantID,
			DeploymentMode: "onpremise"}}, Diagnostics: recovery,
		Protection: protection, Anchors: &httpRecoveryAnchorFixture{},
		RecoverySets: recovery, Restores: recovery,
		Operations: recovery, Traces: recovery, Alerts: recovery,
		PermissionCache: recovery,
		Report:          func(error) {}})
	return handler, access, protection
}
