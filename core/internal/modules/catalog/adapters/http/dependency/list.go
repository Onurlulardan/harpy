package dependencyhttp

import (
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

func (handler *Handler) ListRoute() Route {
	return cataloghttp.NewRoute(http.MethodGet,
		"/api/catalog/packages/{package_id}/versions/{version}/dependencies", handler.list)
}

func (handler *Handler) list(writer http.ResponseWriter, request *http.Request) {
	id, err := cataloghttp.PathPackageID(request, "package_id")
	var version string
	if err == nil {
		version, err = cataloghttp.PathVersion(request)
	}
	if err == nil {
		err = cataloghttp.RequireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.runtime.AccessFor(request, cataloghttp.PermissionRead,
		"catalog.package_version_dependency.list", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	items, err := handler.app.ListPackageVersionDependencies(
		request.Context(), access, id, version,
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.runtime.Success(writer, request, http.StatusOK, struct {
		Items []cataloghttp.PackageDependency `json:"items"`
	}{Items: items})
}
