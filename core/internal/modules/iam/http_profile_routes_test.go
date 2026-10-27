package iam

import (
	"net/http"
	"testing"
)

func TestProfileHTTPRoutesAreExactSeven(t *testing.T) {
	want := []string{
		http.MethodGet + " /api/profile",
		http.MethodPut + " /api/profile",
		http.MethodGet + " /api/iam/profile-schemas/{resource}/active",
		http.MethodGet + " /api/iam/profile-schemas/{resource}/{schema_id}",
		http.MethodPost + " /api/iam/profile-schemas/{resource}/publish",
		http.MethodPost + " /api/iam/profile-schemas/{resource}/backfills/{plan_id}/recover",
		http.MethodPost + " /api/iam/profile-schemas/{resource}/fields/{field_id}/retire",
	}
	routes := (&ProfileHTTP{}).Routes()
	if len(routes) != len(want) {
		t.Fatalf("route count=%d want=%d", len(routes), len(want))
	}
	seen := make(map[string]struct{}, len(routes))
	for index, route := range routes {
		actual := route.Method + " " + route.Pattern
		if actual != want[index] || route.Handler == nil {
			t.Fatalf("route[%d]=%q handler=%v want=%q", index, actual, route.Handler, want[index])
		}
		if _, duplicate := seen[actual]; duplicate {
			t.Fatalf("duplicate route %q", actual)
		}
		seen[actual] = struct{}{}
	}
	var handler *ProfileHTTP
	if routes := handler.Routes(); routes != nil {
		t.Fatalf("nil handler routes=%v", routes)
	}
}

func TestProfileHTTPActiveLiteralWinsOverSchemaParameter(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	fixture.app.results["GetActiveSchema"] = ProfileSchemaView{
		ID: profileHTTPTestSchemaID, Version: 1,
	}
	response := fixture.serve(fixture.request(
		http.MethodGet, "/api/iam/profile-schemas/user/active", "",
	))
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 ||
		fixture.app.calls[0].operation != "GetActiveSchema" {
		t.Fatalf("status=%d calls=%+v body=%s", response.Code, fixture.app.calls, response.Body)
	}
}
