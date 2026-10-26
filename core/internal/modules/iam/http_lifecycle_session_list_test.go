package iam

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func TestLifecycleHTTPSessionListEmitsCanonicalEnvelope(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	portalID := uuid.New()
	created := auth.now.Add(-time.Hour).UTC()
	app := &lifecycleHTTPTestApplication{sessions: []lifecycle.Session{{
		ID: uuid.New(), PortalBindingID: portalID, AccessMode: "portal",
		DeviceKind: "desktop", BrowserFamily: "chromium", OSFamily: "windows",
		Revision: 4, CreatedAt: created, LastSeenAt: created.Add(time.Minute),
		AbsoluteExpiry: created.Add(24 * time.Hour),
	}}}
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
	response := httptest.NewRecorder()
	handler.listSessions(response, request)
	document := decodeAuthResponse(t, response)
	items, ok := document["items"].([]any)
	if response.Code != 200 || !ok || len(items) != 1 || app.listSessionCalls != 1 {
		t.Fatalf("status/items/calls=%d/%v/%d", response.Code, document["items"], app.listSessionCalls)
	}
	item, ok := items[0].(map[string]any)
	if !ok || len(item) != 10 || item["id"] != app.sessions[0].ID.String() ||
		item["portal_binding_id"] != portalID.String() || item["access_mode"] != "portal" ||
		item["device_kind"] != "desktop" || item["browser_family"] != "chromium" ||
		item["os_family"] != "windows" || item["revision"] != float64(4) ||
		item["created_at"] != created.Format(time.RFC3339Nano) {
		t.Fatalf("non-canonical session item: %v", item)
	}
}

func TestLifecycleHTTPSessionListOmitsAbsentPortalBinding(t *testing.T) {
	response := lifecycleSessionListView([]lifecycle.Session{{ID: uuid.New()}})
	if len(response.Items) != 1 || response.Items[0].PortalBinding != nil {
		t.Fatalf("unexpected portal binding: %+v", response.Items)
	}
}
