package supporthttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProtectedRoutesRequestExactSupportAuthority(t *testing.T) {
	tests := []struct {
		method, pattern, scope, permission, operation string
	}{
		{http.MethodGet, "/api/admin/contact-messages", "platform", "platform.contact_messages.read", "support.contact.list"},
		{http.MethodGet, "/api/admin/contact-messages/stats", "platform", "platform.contact_messages.read", "support.contact.stats"},
		{http.MethodGet, "/api/admin/contact-messages/{id}", "platform", "platform.contact_messages.read", "support.contact.read"},
		{http.MethodPut, "/api/admin/contact-messages/{id}", "platform", "platform.contact_messages.manage", "support.contact.update"},
		{http.MethodPost, "/api/bug-reports", "reporter", "bug_reports.create", "support.bug.create"},
		{http.MethodGet, "/api/bug-reports/mine", "reporter", "bug_reports.read_own", "support.bug.mine"},
		{http.MethodGet, "/api/bug-reports/{id}", "reporter", "bug_reports.read_own", "support.bug.read"},
		{http.MethodGet, "/api/bug-reports/admin", "tenant", "bug_reports.read", "support.bug.tenant.list"},
		{http.MethodGet, "/api/bug-reports/admin/stats", "tenant", "bug_reports.read", "support.bug.tenant.stats"},
		{http.MethodGet, "/api/bug-reports/admin/{id}", "tenant", "bug_reports.read", "support.bug.tenant.read"},
		{http.MethodPut, "/api/bug-reports/admin/{id}", "tenant", "bug_reports.manage", "support.bug.tenant.update"},
		{http.MethodGet, "/api/admin/bug-reports", "platform", "platform.bug_reports.read", "support.bug.platform.list"},
		{http.MethodGet, "/api/admin/bug-reports/stats", "platform", "platform.bug_reports.read", "support.bug.platform.stats"},
		{http.MethodGet, "/api/admin/bug-reports/{id}", "platform", "platform.bug_reports.read", "support.bug.platform.read"},
		{http.MethodPut, "/api/admin/bug-reports/{id}", "platform", "platform.bug_reports.manage", "support.bug.platform.update"},
	}
	for _, test := range tests {
		t.Run(test.method+" "+test.pattern, func(t *testing.T) {
			handler, _, access := handlerFixture()
			route := findRoute(t, handler.Routes(), test.method, test.pattern)
			request := authorityRequest(test.method, test.pattern)
			route.Handler.ServeHTTP(httptest.NewRecorder(), request)
			if string(access.last.Scope) != test.scope || access.last.Permission != test.permission ||
				access.last.Operation != test.operation {
				t.Fatalf("authority = %s/%s/%s, want %s/%s/%s", access.last.Scope,
					access.last.Permission, access.last.Operation, test.scope, test.permission, test.operation)
			}
		})
	}
}

func findRoute(t *testing.T, routes []Route, method, pattern string) Route {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Pattern == pattern {
			return route
		}
	}
	t.Fatalf("route %s %s not found", method, pattern)
	return Route{}
}

func authorityRequest(method, pattern string) *http.Request {
	body := ""
	if method == http.MethodPost {
		body = `{"title":"Broken page","description":"The page cannot be opened"}`
	}
	if method == http.MethodPut {
		body = `{"status":"resolved","expected_revision":1}`
	}
	target := strings.Replace(pattern, "{id}", "01900000-0000-7000-8000-000000000001", 1)
	if strings.HasPrefix(pattern, "/api/admin/bug-reports/{id}") {
		target += "?tenant_id=10000000-0000-4000-8000-000000000001"
	}
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.SetPathValue("id", "01900000-0000-7000-8000-000000000001")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "support-command-0001")
	}
	return request
}
