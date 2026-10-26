package lifecyclehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

func (app *installationApplication) GetDemoStatus(
	context.Context, cataloghttp.Access, uuid.UUID,
) (cataloghttp.DemoStatus, error) {
	app.called = "demo_status"
	return cataloghttp.DemoStatus{}, nil
}

func (app *installationApplication) PlanDemoRemoval(
	context.Context, cataloghttp.Access, uuid.UUID,
) (cataloghttp.DemoRemovalPlan, error) {
	app.called = "demo_plan"
	return cataloghttp.DemoRemovalPlan{}, nil
}

func (app *installationApplication) InstallDemo(
	_ context.Context, _ cataloghttp.Access, _ uuid.UUID, key string,
) (cataloghttp.Operation, error) {
	app.called, app.key = "demo_install", key
	return cataloghttp.Operation{}, nil
}

func (app *installationApplication) RemoveDemo(
	_ context.Context, _ cataloghttp.Access, _ uuid.UUID, expected, key string,
) (cataloghttp.Operation, error) {
	app.called, app.expectedPlan, app.key = "demo_remove", expected, key
	return cataloghttp.Operation{}, nil
}

func TestDemoLifecycleRoutesUseExactPermissionAndDurableMutationStatus(t *testing.T) {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name, call, operation string
		permission            cataloghttp.Permission
		status                int
		route                 func(*Handler) Route
		body                  string
	}{
		{"status", "demo_status", "catalog.installation.demo.get",
			PermissionRead, 200, (*Handler).GetDemoStatusRoute, ""},
		{"plan", "demo_plan", "catalog.installation.demo.removal_plan",
			PermissionDemoManage, 200, (*Handler).PlanDemoRemovalRoute, ""},
		{"install", "demo_install", "catalog.installation.demo.install",
			PermissionDemoManage, 202, (*Handler).InstallDemoRoute, ""},
		{"remove", "demo_remove", "catalog.installation.demo.remove",
			PermissionDemoManage, 202, (*Handler).RemoveDemoRoute,
			`{"expected_plan_sha256":"` + digest + `"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			access := new(installationAccess)
			app := new(installationApplication)
			runtime, err := cataloghttp.NewRuntime(
				access, func(error) {}, func(*http.Request) string {
					return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			handler := New(runtime, app)
			route := test.route(handler)
			installationID := uuid.MustParse("018f6000-0000-7000-8000-000000000001")
			path := strings.ReplaceAll(route.Pattern, "{installation_id}", installationID.String())
			request := httptest.NewRequest(route.Method, path, strings.NewReader(test.body))
			request.SetPathValue("installation_id", installationID.String())
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			if route.Method == http.MethodPost || route.Method == http.MethodDelete {
				request.Header.Set("Idempotency-Key", "catalog-demo-key-0001")
			}
			response := httptest.NewRecorder()
			route.Handler.ServeHTTP(response, request)
			if response.Code != test.status || app.called != test.call ||
				access.operation != test.operation || access.permission != test.permission ||
				test.call == "demo_remove" && app.expectedPlan != digest {
				t.Fatalf("status=%d call=%s operation=%s permission=%s plan=%s body=%s",
					response.Code, app.called, access.operation, access.permission,
					app.expectedPlan, response.Body.String())
			}
		})
	}
}
