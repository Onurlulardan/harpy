package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/google/uuid"
)

func TestPostRestoreOwnerReceivesNormalSessionGrant(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.loginGrant = fixture.newGrant(t, fixture.claims)
	handler := &PostRestoreHTTP{auth: fixture.handler}
	request := httptest.NewRequest(http.MethodPost,
		"/api/auth/post-restore/reactivate-owner", nil)
	request.Header.Set("Origin", fixture.origin)
	password := []byte("new-owner-password")
	response, cookies, err := handler.issueRestoredOwnerSession(request,
		postrestore.ReactivateResult{
			UserID: uuid.New(), TenantID: uuid.New(), TenantSlug: "tenant-a",
			Identifier: "owner@example.com", SessionState: "active",
		}, password, fixture.origin)
	if err != nil {
		t.Fatal(err)
	}
	if response.Token == "" || response.ExpiresAt.IsZero() || len(cookies) != 2 ||
		fixture.app.loginCalls != 1 ||
		fixture.app.loginInput.Mode != AuthHTTPLoginOperator ||
		fixture.app.loginInput.TenantSlug != "tenant-a" ||
		fixture.app.loginInput.Identifier != "owner@example.com" {
		t.Fatalf("response=%+v cookies=%v input=%+v",
			response, cookies, fixture.app.loginInput)
	}
	for _, value := range password {
		if value == 0 {
			t.Fatal("caller password was unexpectedly mutated")
		}
	}
}
