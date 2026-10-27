package iam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func TestLifecycleHTTPPortalSetupReadEmitsSecretFreeRevisionedView(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	bindingID := uuid.New()
	userID := uuid.New()
	completedAt := auth.now.Add(-3 * time.Minute).UTC()
	submittedAt := completedAt.Add(time.Minute)
	decisionAt := submittedAt.Add(time.Minute)
	updatedAt := decisionAt.Add(time.Minute)
	app := &lifecycleHTTPTestApplication{portalSetup: lifecycle.PortalSetup{
		BindingID: bindingID, UserID: userID, SetupState: "completed",
		ApprovalState: "approved", Revision: 2,
		SetupCompletedAt: completedAt, SubmittedAt: submittedAt,
		DecisionAt: decisionAt, UpdatedAt: updatedAt,
	}}
	handler, err := newLifecycleHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := lifecycleProtectedRequest(token, auth.claims.Tenant())
	request.SetPathValue("binding_id", bindingID.String())
	response := httptest.NewRecorder()
	handler.readPortalSetup(response, request)
	document := decodeAuthResponse(t, response)
	if response.Code != http.StatusOK ||
		response.Header().Get("ETag") != `"revision-2"` ||
		app.readPortalSetupCalls != 1 || app.readPortalSetupID != bindingID ||
		len(document) != 9 || document["binding_id"] != bindingID.String() ||
		document["user_id"] != userID.String() ||
		document["setup_state"] != "completed" ||
		document["approval_state"] != "approved" ||
		document["revision"] != float64(2) ||
		document["updated_at"] != updatedAt.Format(time.RFC3339Nano) {
		t.Fatalf("status=%d etag=%q calls=%d body=%s", response.Code,
			response.Header().Get("ETag"), app.readPortalSetupCalls, response.Body.String())
	}
	for _, forbidden := range [][]byte{
		[]byte("password"), []byte("token"), []byte("email"),
		[]byte("source_evidence"), []byte("decision_actor"), []byte("provenance"),
	} {
		if bytes.Contains(bytes.ToLower(response.Body.Bytes()), forbidden) {
			t.Fatalf("portal setup view exposed %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestLifecycleHTTPPortalSetupReadRejectsSubstitutedBinding(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	requested := uuid.New()
	app := &lifecycleHTTPTestApplication{portalSetup: lifecycle.PortalSetup{
		BindingID: uuid.New(), UserID: uuid.New(), SetupState: "required",
		ApprovalState: "not_submitted", Revision: 1, UpdatedAt: auth.now,
	}}
	handler, err := newLifecycleHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := lifecycleProtectedRequest(token, auth.claims.Tenant())
	request.SetPathValue("binding_id", requested.String())
	response := httptest.NewRecorder()
	handler.readPortalSetup(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
