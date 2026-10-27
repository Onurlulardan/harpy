package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRBACHTTPRequiresExactBearerAndTenant(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	valid := fixture.request(http.MethodGet, "/api/permissions", "")
	response := httptest.NewRecorder()
	fixture.handler.listPermissions(response, valid)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("valid request = %d / calls=%d", response.Code, len(fixture.app.calls))
	}
	wrong := fixture.request(http.MethodGet, "/api/permissions", "")
	wrong.Header.Set(authTenantIDHeader, uuid.New().String())
	response = httptest.NewRecorder()
	fixture.handler.listPermissions(response, wrong)
	if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 1 {
		t.Fatalf("tenant substitution = %d / calls=%d", response.Code, len(fixture.app.calls))
	}
	duplicate := fixture.request(http.MethodGet, "/api/permissions", "")
	duplicate.Header.Add("Authorization", "Bearer "+fixture.token)
	response = httptest.NewRecorder()
	fixture.handler.listPermissions(response, duplicate)
	if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 1 {
		t.Fatalf("duplicate bearer = %d / calls=%d", response.Code, len(fixture.app.calls))
	}
}

func TestRBACCheckRejectsBodyAuthorityAndWildcards(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	bodies := []string{
		`{"resource_slug":"iam.users","action_slug":"read","user_id":"` + uuid.New().String() + `"}`,
		`{"resource_slug":"iam.users","action_slug":"read","tenant_id":"` + uuid.New().String() + `"}`,
		`{"resource_slug":"ALL","action_slug":"read"}`,
	}
	for _, body := range bodies {
		response := httptest.NewRecorder()
		fixture.handler.check(response, fixture.request(http.MethodPost, "/api/permissions/check", body))
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("body=%s status=%d calls=%d", body, response.Code, len(fixture.app.calls))
		}
	}
	valid := `{"resource_slug":"iam.users","action_slug":"read"}`
	response := httptest.NewRecorder()
	fixture.handler.check(response, fixture.request(http.MethodPost, "/api/permissions/check", valid))
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 ||
		fixture.app.calls[0] != "Check" {
		t.Fatalf("valid check = %d / calls=%v", response.Code, fixture.app.calls)
	}
}

func TestRBACBatchCheckRejectsNestedAuthorityAndDuplicateDecisions(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	bodies := []string{
		`{"checks":[{"resource_slug":"iam.users","action_slug":"read","user_id":"` + uuid.New().String() + `"}]}`,
		`{"checks":[{"resource_slug":"iam.users","action_slug":"ALL"}]}`,
		`{"checks":[{"resource_slug":"iam.users","action_slug":"read"},{"resource_slug":"iam.users","action_slug":"read"}]}`,
	}
	for _, body := range bodies {
		response := httptest.NewRecorder()
		fixture.handler.batchCheck(
			response, fixture.request(http.MethodPost, "/api/permissions/batch-check", body),
		)
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("body=%s status=%d calls=%d", body, response.Code, len(fixture.app.calls))
		}
	}
}

func TestRBACHTTPRejectsPortalBearer(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	app := &rbacHTTPTestApplication{}
	handler, err := newRBACHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	portal := auth.newClaims(t, true)
	token, err := auth.security.accessSigner.Sign(portal)
	if err != nil {
		t.Fatal(err)
	}
	request := invitationProtectedRequest(
		http.MethodGet, "/api/permissions", "", token, portal.Tenant(),
	)
	response := httptest.NewRecorder()
	handler.listPermissions(response, request)
	if response.Code != http.StatusUnauthorized || len(app.calls) != 0 {
		t.Fatalf("portal bearer=%d calls=%d", response.Code, len(app.calls))
	}
}

func TestRBACCacheEndpointsAreExplicitlyGone(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	response := httptest.NewRecorder()
	fixture.handler.cacheRetired(
		response, fixture.request(http.MethodGet, "/api/permissions/cache/stats", ""),
	)
	if response.Code != http.StatusGone ||
		!strings.Contains(response.Body.String(), "PERMISSION_CACHE_RETIRED") ||
		len(fixture.app.calls) != 0 {
		t.Fatalf("cache response = %d / %s", response.Code, response.Body.String())
	}
}

func TestRBACCacheRetirementStillRequiresAuthentication(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/api/permissions/cache/stats", nil)
	response := httptest.NewRecorder()
	fixture.handler.cacheRetired(response, request)
	if response.Code != http.StatusUnauthorized ||
		strings.Contains(response.Body.String(), "PERMISSION_CACHE_RETIRED") {
		t.Fatalf("unauthenticated cache response=%d body=%s", response.Code, response.Body.String())
	}
}
