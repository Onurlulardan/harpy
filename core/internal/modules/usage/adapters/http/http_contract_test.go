package usagehttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func TestRoutesExposeExactLegacyUsageSurface(t *testing.T) {
	handler, _, _ := testHandler(t)
	got := make([]string, 0, 16)
	for _, route := range handler.Routes() {
		got = append(got, route.Method+" "+route.Pattern)
	}
	want := []string{
		"GET /api/usage-analytics/overview", "GET /api/usage-analytics/users/active",
		"GET /api/usage-analytics/users/top",
		"GET /api/usage-analytics/users/{user_id}/summary",
		"GET /api/usage-analytics/users/{user_id}/timeline",
		"GET /api/usage-analytics/features", "GET /api/usage-analytics/heatmap",
		"GET /api/usage-analytics/trend", "GET /api/usage-analytics/tenant/overview",
		"GET /api/usage-analytics/tenant/users/active",
		"GET /api/usage-analytics/tenant/users/top",
		"GET /api/usage-analytics/tenant/users/{user_id}/summary",
		"GET /api/usage-analytics/tenant/users/{user_id}/timeline",
		"GET /api/usage-analytics/tenant/features",
		"GET /api/usage-analytics/tenant/heatmap", "GET /api/usage-analytics/tenant/trend",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("route drift:\ngot  %v\nwant %v", got, want)
	}
}

func TestTenantOverviewRequestsExactScopePermissionAndOperation(t *testing.T) {
	handler, verifier, _ := testHandler(t)
	request := httptest.NewRequest(http.MethodGet,
		"/api/usage-analytics/tenant/overview", nil)
	request.Header.Set("Authorization", "Bearer opaque")
	request.Header.Set("X-Tenant-ID", httpTenant.String())
	response := httptest.NewRecorder()
	testMux(t, handler).ServeHTTP(response, request)
	if response.Code != http.StatusOK || verifier.last.Scope != domain.TenantScope ||
		verifier.last.Permission != application.PermissionTenantRead ||
		verifier.last.Operation != application.OperationTenantOverview {
		t.Fatalf("wrong tenant request contract: status=%d access=%+v",
			response.Code, verifier.last)
	}
}

func TestInvalidTimeRangeIsStable422AndDoesNotCallApplication(t *testing.T) {
	handler, _, app := testHandler(t)
	request := httptest.NewRequest(http.MethodGet,
		"/api/usage-analytics/overview?start_time=not-a-time", nil)
	response := httptest.NewRecorder()
	testMux(t, handler).ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity || app.overviewCalls != 0 ||
		!strings.Contains(response.Body.String(), `"code":"INPUT_INVALID"`) {
		t.Fatalf("invalid query contract drift: status=%d calls=%d body=%s",
			response.Code, app.overviewCalls, response.Body.String())
	}
}

func TestPlatformUserResponseContainsOnlyPseudonymousSubject(t *testing.T) {
	handler, verifier, app := testHandler(t)
	pseudonym := strings.Repeat("a", 64)
	request := httptest.NewRequest(http.MethodGet,
		"/api/usage-analytics/users/"+pseudonym+"/summary", nil)
	response := httptest.NewRecorder()
	testMux(t, handler).ServeHTTP(response, request)
	var payload map[string]any
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &payload) != nil ||
		app.summaryUser != pseudonym || verifier.last.Scope != domain.PlatformScope ||
		verifier.last.Permission != application.PermissionPlatformRead {
		t.Fatalf("platform pseudonym contract drift: status=%d body=%s",
			response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), httpActor.String()) ||
		strings.Contains(response.Body.String(), httpTenant.String()) {
		t.Fatal("platform response exposed raw identity")
	}
}
