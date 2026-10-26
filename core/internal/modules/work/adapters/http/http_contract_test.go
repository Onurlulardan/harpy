package workhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fixtureAccess struct{}

func (fixtureAccess) Verify(_ context.Context, request AccessRequest) (Access, error) {
	consumer, build := "", ""
	if request.Compatibility {
		consumer, build = "frontend-v3", "build-1"
	}
	return NewAccess(uuid.New(), uuid.New(), 1, 1, request.Authority,
		request.Operation, consumer, build, request.Compatibility)
}

type fixtureApplication struct{ hits int }

func (*fixtureApplication) List(context.Context, ListInput) (json.RawMessage, error) {
	return json.RawMessage(`{"items":[]}`), nil
}
func (*fixtureApplication) Get(_ context.Context, input GetInput) (json.RawMessage, error) {
	return json.Marshal(map[string]string{"id": input.IDs["board_id"]})
}
func (*fixtureApplication) Write(context.Context, WriteInput) (json.RawMessage, error) {
	return json.RawMessage(`{"row_version":2}`), nil
}
func (application *fixtureApplication) RecordCompatibilityHit(
	context.Context, CompatibilityHit,
) error {
	application.hits++
	return nil
}

func fixtureHandler(t *testing.T) (*Handler, *fixtureApplication) {
	t.Helper()
	application := &fixtureApplication{}
	handler, err := NewHandler(HandlerDependencies{Access: fixtureAccess{},
		Application: application,
		Report:      func(error) {}, Correlation: func(*http.Request) string { return "request-1" }})
	if err != nil {
		t.Fatal(err)
	}
	return handler, application
}

func TestRoutesCoverCanonicalAndFinalCompatibilityEdges(t *testing.T) {
	handler, _ := fixtureHandler(t)
	seen := map[string]bool{}
	for _, route := range handler.Routes() {
		key := route.Method + " " + route.Pattern
		if seen[key] {
			t.Fatalf("duplicate route %s", key)
		}
		seen[key] = true
	}
	if len(seen) != 68 {
		t.Fatalf("Work route count = %d, want 68", len(seen))
	}
	for _, key := range []string{
		"GET /api/work/boards", "POST /api/work/tasks/{task_id}/moves",
		"GET /api/work/inbox", "GET /api/work/calendar",
		"GET /api/workflow-service/boards",
		"GET /api/workflow-service/board-tasks",
		"GET /api/workflow-service/calendar-events",
		"GET /api/workflow-service/boards/available-users",
		"GET /api/workflow-service/board-tasks/my-tasks",
		"PUT /api/workflow-service/boards/{board_id}/columns/reorder",
		"POST /api/workflow-service/boards/{board_id}/tasks",
		"DELETE /api/workflow-service/board-tasks/{task_id}/comments/{comment_id}",
	} {
		if !seen[key] {
			t.Fatalf("missing route %s", key)
		}
	}
}

func TestCompatibilityRouteInventoryIsExact(t *testing.T) {
	want := []string{
		"GET /api/workflow-service/boards/available-users",
		"GET /api/workflow-service/board-tasks/my-tasks",
		"GET /api/workflow-service/board-tasks/calendar",
		"GET /api/workflow-service/boards", "POST /api/workflow-service/boards",
		"GET /api/workflow-service/boards/{board_id}",
		"PUT /api/workflow-service/boards/{board_id}",
		"DELETE /api/workflow-service/boards/{board_id}",
		"GET /api/workflow-service/boards/{board_id}/members",
		"POST /api/workflow-service/boards/{board_id}/members",
		"PUT /api/workflow-service/boards/{board_id}/members/{member_id}",
		"DELETE /api/workflow-service/boards/{board_id}/members/{member_id}",
		"GET /api/workflow-service/boards/{board_id}/my-role",
		"POST /api/workflow-service/boards/{board_id}/columns",
		"PUT /api/workflow-service/boards/{board_id}/columns/reorder",
		"PUT /api/workflow-service/boards/{board_id}/columns/{column_id}",
		"DELETE /api/workflow-service/boards/{board_id}/columns/{column_id}",
		"POST /api/workflow-service/boards/{board_id}/tasks",
		"GET /api/workflow-service/board-tasks", "POST /api/workflow-service/board-tasks",
		"GET /api/workflow-service/board-tasks/{task_id}",
		"PUT /api/workflow-service/board-tasks/{task_id}",
		"DELETE /api/workflow-service/board-tasks/{task_id}",
		"PUT /api/workflow-service/board-tasks/{task_id}/move",
		"GET /api/workflow-service/board-tasks/{task_id}/comments",
		"POST /api/workflow-service/board-tasks/{task_id}/comments",
		"DELETE /api/workflow-service/board-tasks/{task_id}/comments/{comment_id}",
		"GET /api/workflow-service/calendar-events",
		"POST /api/workflow-service/calendar-events",
		"GET /api/workflow-service/calendar-events/{event_id}",
		"PUT /api/workflow-service/calendar-events/{event_id}",
		"DELETE /api/workflow-service/calendar-events/{event_id}",
	}
	routes := compatibilityRoutes()
	if len(routes) != len(want) {
		t.Fatalf("compatibility routes = %d, want %d", len(routes), len(want))
	}
	for index, route := range routes {
		got := route.method + " " + route.pattern
		if got != want[index] || !route.compatibility {
			t.Fatalf("route %d = %q compatibility=%t, want %q", index, got,
				route.compatibility, want[index])
		}
	}
}

func TestGetUsesVerifiedAccessAndCanonicalEnvelope(t *testing.T) {
	handler, _ := fixtureHandler(t)
	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	id := uuid.New().String()
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/work/boards/"+id, nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), id) ||
		!strings.Contains(response.Body.String(), `"success":true`) {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}

func TestCompatibilityHitIsPersistedBeforeRead(t *testing.T) {
	handler, application := fixtureHandler(t)
	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	id := uuid.New().String()
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet,
		"/api/workflow-service/boards/"+id, nil))
	if response.Code != http.StatusOK || application.hits != 1 {
		t.Fatalf("compatibility response=%d hits=%d", response.Code, application.hits)
	}
}
