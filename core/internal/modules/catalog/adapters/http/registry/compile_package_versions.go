package registryhttp

import "net/http"

func (handler *Handler) CompilePackageVersionsRoute() Route {
	return route(http.MethodGet, "/api/catalog/packages/{package_id}/versions", handler.compilePackageVersions)
}

func (handler *Handler) compilePackageVersions(writer http.ResponseWriter, request *http.Request) {
	id, err := pathPackageID(request, "package_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.package_version.list", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	items, err := handler.app.ListPackageVersions(request.Context(), access, id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, struct {
		Items []PackageVersion `json:"items"`
	}{Items: items})
}
