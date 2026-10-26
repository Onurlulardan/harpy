package registryhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type packageVersionAccess struct{ operation string }

func (access *packageVersionAccess) Verify(_ context.Context,
	request cataloghttp.AccessRequest,
) (cataloghttp.Access, error) {
	access.operation = request.Operation
	return cataloghttp.NewAccess(uuid.New(), uuid.New(), 1, 1, 1, time.Time{},
		request.Permission, request.Operation)
}

type packageVersionApplication struct {
	*packageApplication
	called string
}

func (app *packageVersionApplication) ListPackageVersions(context.Context, cataloghttp.Access,
	uuid.UUID,
) ([]cataloghttp.PackageVersion, error) {
	app.called = "list_package_versions"
	return []cataloghttp.PackageVersion{}, nil
}

func (app *packageVersionApplication) GetPackageVersion(context.Context, cataloghttp.Access,
	uuid.UUID, string,
) (cataloghttp.PackageVersion, error) {
	app.called = "get_package_version"
	return cataloghttp.PackageVersion{}, nil
}

func exercisePackageVersion(expectedCall, expectedOperation string,
	selectRoute func(*Handler) Route,
) error {
	access := new(packageVersionAccess)
	application := &packageVersionApplication{packageApplication: new(packageApplication)}
	runtime, err := cataloghttp.NewRuntime(access, func(error) {}, func(*http.Request) string {
		return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
	})
	if err != nil {
		return err
	}
	handler := New(runtime, application)
	route := selectRoute(handler)
	id := uuid.MustParse("e3000000-0000-0000-0000-000000000001")
	path := strings.ReplaceAll(route.Pattern, "{package_id}", id.String())
	path = strings.ReplaceAll(path, "{version}", "1.2.3")
	request := httptest.NewRequest(route.Method, path, nil)
	request.SetPathValue("package_id", id.String())
	request.SetPathValue("version", "1.2.3")
	response := httptest.NewRecorder()
	route.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || application.called != expectedCall ||
		access.operation != expectedOperation {
		return fmt.Errorf("package version route status=%d call=%s operation=%s",
			response.Code, application.called, access.operation)
	}
	return nil
}
