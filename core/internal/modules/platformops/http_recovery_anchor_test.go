package platformops

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type recoveryAnchorHTTPFixture struct {
	listSubject  platformopsapi.RecoveryAnchorSubject
	options      platformopsapi.RecoveryAnchorOptionsCommand
	registration platformopsapi.RecoveryAnchorRegisterCommand
	deletion     platformopsapi.RecoveryAnchorDeleteCommand
}

func (fixture *recoveryAnchorHTTPFixture) ListRecoveryAnchors(
	_ context.Context, subject platformopsapi.RecoveryAnchorSubject,
) ([]platformopsapi.RecoveryAnchor, error) {
	fixture.listSubject = subject
	return []platformopsapi.RecoveryAnchor{{AnchorID: httpOperationID, Generation: 1,
		CredentialID: []byte{1, 2, 3}, RPID: "localhost",
		UserVerification: "required", State: "active"}}, nil
}

func (fixture *recoveryAnchorHTTPFixture) CreateRecoveryAnchorOptions(
	_ context.Context, command platformopsapi.RecoveryAnchorOptionsCommand,
) (platformopsapi.RecoveryAnchorRegistrationOptions, error) {
	fixture.options = command
	return platformopsapi.RecoveryAnchorRegistrationOptions{OperationID: command.OperationID,
		Challenge: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		RPID:      "localhost", UserID: "AQID"}, nil
}

func (fixture *recoveryAnchorHTTPFixture) RegisterRecoveryAnchor(
	_ context.Context, command platformopsapi.RecoveryAnchorRegisterCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	fixture.registration = command
	return platformopsapi.RecoveryAnchorOperation{OperationID: command.OperationID,
		Kind: "anchor_register", State: "accepted"}, nil
}

func (fixture *recoveryAnchorHTTPFixture) DeleteRecoveryAnchor(
	_ context.Context, command platformopsapi.RecoveryAnchorDeleteCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	fixture.deletion = command
	return platformopsapi.RecoveryAnchorOperation{OperationID: command.OperationID,
		Kind: "anchor_delete", State: "accepted"}, nil
}

func recoveryAnchorHandlerFixture(
	t *testing.T,
) (*HTTPHandler, *httpAccessFixture, *recoveryAnchorHTTPFixture) {
	t.Helper()
	access := &httpAccessFixture{value: HTTPAccess{ActorID: httpActorID,
		TenantID: httpTenantID, SessionID: httpSessionID,
		TokenVersion: 1, AuthorizationVersion: 1, PlatformClock: 1}}
	anchors := &recoveryAnchorHTTPFixture{}
	recovery := &httpRecoveryFixture{}
	handler, err := NewHTTPHandler(HTTPDependencies{Access: access,
		Environment: httpEnvironmentFixture{value: HTTPEnvironment{
			EnvironmentID: httpEnvironmentID, TenantID: httpTenantID,
			DeploymentMode: "onpremise"}}, Diagnostics: recovery,
		Protection: &httpProtectionFixture{}, Anchors: anchors,
		RecoverySets: recovery, Restores: recovery, Operations: recovery,
		Traces: recovery, Alerts: recovery, PermissionCache: recovery,
		Report: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	return handler, access, anchors
}

func TestRecoveryAnchorListUsesOwnerCapability(t *testing.T) {
	handler, access, fixture := recoveryAnchorHandlerFixture(t)
	response := httptest.NewRecorder()
	handler.listRecoveryAnchors(response,
		httptest.NewRequest(http.MethodGet, "/anchors", nil))
	if response.Code != http.StatusOK || fixture.listSubject.ActorID != httpActorID ||
		len(access.requests) != 1 ||
		access.requests[0].Capability != "platform.recovery.anchor.read" ||
		access.requests[0].RequireStepUp {
		t.Fatalf("anchor list drifted: %d %#v", response.Code, access.requests)
	}
}

func TestRecoveryAnchorOptionsAreWebAuthnAndStable(t *testing.T) {
	handler, access, fixture := recoveryAnchorHandlerFixture(t)
	request := httptest.NewRequest(http.MethodPost, "/anchor-options", nil)
	request.Header.Set("Idempotency-Key", "anchor-options-0001")
	response := httptest.NewRecorder()
	handler.createRecoveryAnchorOptions(response, request)
	if response.Code != http.StatusOK || fixture.options.OperationID == "" ||
		!strings.Contains(response.Body.String(), `"authenticatorAttachment":"cross-platform"`) ||
		!strings.Contains(response.Body.String(), `"id":"localhost"`) {
		t.Fatalf("anchor options=%d %s", response.Code, response.Body.String())
	}
	first := fixture.options.OperationID
	handler.createRecoveryAnchorOptions(httptest.NewRecorder(), request.Clone(request.Context()))
	if fixture.options.OperationID != first || !access.requests[0].RequireStepUp {
		t.Fatal("anchor options identity or step-up drifted")
	}
}

func TestRecoveryAnchorRegisterAndDeleteUseExactEvidence(t *testing.T) {
	handler, access, fixture := recoveryAnchorHandlerFixture(t)
	body := `{"operation_id":"` + httpOperationID + `","credential":{"id":"key"}}`
	register := httptest.NewRequest(http.MethodPost, "/anchors", bytes.NewBufferString(body))
	register.Header.Set("Content-Type", "application/json")
	register.Header.Set("Idempotency-Key", "anchor-register-0001")
	registered := httptest.NewRecorder()
	handler.registerRecoveryAnchor(registered, register)
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/anchors/"+httpOperationID, nil)
	deleteRequest.SetPathValue("anchor_id", httpOperationID)
	deleteRequest.Header.Set("Idempotency-Key", "anchor-delete-00001")
	deleted := httptest.NewRecorder()
	handler.deleteRecoveryAnchor(deleted, deleteRequest)
	if registered.Code != http.StatusAccepted || deleted.Code != http.StatusAccepted ||
		fixture.registration.OperationID != httpOperationID ||
		fixture.deletion.AnchorID != httpOperationID ||
		len(fixture.registration.CredentialJSON) == 0 || len(access.requests) != 2 ||
		!access.requests[0].RequireStepUp || !access.requests[1].RequireStepUp {
		t.Fatalf("anchor mutations drifted: %d %d", registered.Code, deleted.Code)
	}
}
