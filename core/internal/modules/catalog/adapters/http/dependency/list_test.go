package dependencyhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type accessVerifier struct{ operation string }

func (verifier *accessVerifier) Verify(
	_ context.Context, request cataloghttp.AccessRequest,
) (cataloghttp.Access, error) {
	verifier.operation = request.Operation
	return cataloghttp.NewAccess(uuid.New(), uuid.New(), 1, 1, 1, time.Time{},
		request.Permission, request.Operation)
}

type application struct{ called bool }

func (app *application) ListPackageVersionDependencies(
	context.Context, cataloghttp.Access, uuid.UUID, string,
) ([]cataloghttp.PackageDependency, error) {
	app.called = true
	return []cataloghttp.PackageDependency{{
		PackageID:    "e3000000-0000-0000-0000-000000000002",
		VersionRange: "^1.0.0", Kind: "required",
	}}, nil
}

func TestListPackageVersionDependencies(t *testing.T) {
	verifier, app := new(accessVerifier), new(application)
	runtime, err := cataloghttp.NewRuntime(verifier, func(error) {}, func(*http.Request) string {
		return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
	})
	if err != nil {
		t.Fatal(err)
	}
	route := New(runtime, app).ListRoute()
	id := "e3000000-0000-0000-0000-000000000001"
	path := strings.ReplaceAll(route.Pattern, "{package_id}", id)
	path = strings.ReplaceAll(path, "{version}", "1.2.3")
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.SetPathValue("package_id", id)
	request.SetPathValue("version", "1.2.3")
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !app.called ||
		verifier.operation != "catalog.package_version_dependency.list" {
		t.Fatalf("status=%d called=%t operation=%s", response.Code, app.called,
			verifier.operation)
	}
}
