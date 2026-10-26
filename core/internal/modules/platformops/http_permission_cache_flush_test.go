package platformops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type permissionCacheFixture struct {
	httpRecoveryFixture
	request platformopsapi.PermissionCacheFlush
}

func (fixture *permissionCacheFixture) FlushPermissionCache(
	_ context.Context,
	request platformopsapi.PermissionCacheFlush,
) (platformopsapi.PermissionCacheResult, error) {
	fixture.request = request
	return platformopsapi.PermissionCacheResult{
		AuthorizationClock: 9,
		Revision:           3,
	}, nil
}

func TestPermissionCacheFlushRequiresExactStepUpAndStableIdentity(t *testing.T) {
	fixture := &permissionCacheFixture{}
	access := &httpAccessFixture{value: HTTPAccess{
		ActorID: httpActorID, TenantID: httpTenantID, SessionID: httpSessionID,
		TokenVersion: 4, AuthorizationVersion: 5, PlatformClock: 6,
	}}
	protection := &httpProtectionFixture{}
	handler, err := NewHTTPHandler(HTTPDependencies{
		Access: access, Environment: httpEnvironmentFixture{value: HTTPEnvironment{
			EnvironmentID: httpEnvironmentID, TenantID: httpTenantID,
			DeploymentMode: "onpremise",
		}},
		Diagnostics: fixture, Protection: protection,
		Anchors:      &httpRecoveryAnchorFixture{},
		RecoverySets: fixture, Restores: fixture,
		Operations: fixture, Traces: fixture, Alerts: fixture,
		PermissionCache: fixture, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost,
		"/api/platform/operations/permission-cache/flush", nil)
	request.Header.Set("Idempotency-Key", "permission-flush-0001")
	response := httptest.NewRecorder()
	handler.flushPermissionCache(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("permission cache response=%d %s", response.Code, response.Body.String())
	}
	if len(access.requests) != 1 || !access.requests[0].RequireStepUp ||
		access.requests[0].Capability != "platform.operations.permission_cache.flush" {
		t.Fatal("permission cache flush did not require the exact capability")
	}
	wantPurpose, _ := platformopsapi.StepUpPurpose(httpTenantID,
		"platform.operations.permission_cache.flush", httpTenantID)
	if access.requests[0].StepUpPurpose != wantPurpose ||
		fixture.request.TenantID != httpTenantID || fixture.request.ActorID != httpActorID ||
		fixture.request.TokenVersion != 4 || fixture.request.AuthorizationVersion != 5 ||
		fixture.request.PlatformClock != 6 || len(fixture.request.RequestSHA256) != 64 {
		t.Fatalf("permission cache command binding drifted: %#v", fixture.request)
	}
	operationID := fixture.request.OperationID
	handler.flushPermissionCache(httptest.NewRecorder(), request.Clone(request.Context()))
	if fixture.request.OperationID != operationID {
		t.Fatal("permission cache idempotency did not produce stable operation identity")
	}
}

func TestPermissionCacheFlushRejectsMissingIdempotencyKey(t *testing.T) {
	handler, _, _ := httpHandlerFixture()
	request := httptest.NewRequest(http.MethodPost,
		"/api/platform/operations/permission-cache/flush", nil)
	response := httptest.NewRecorder()
	handler.flushPermissionCache(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("missing idempotency key response=%d", response.Code)
	}
}
