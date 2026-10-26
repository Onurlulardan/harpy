package registryhttp

import "net/http"

func (handler *Handler) ResolvePackageVersionRoute() Route {
	return route(http.MethodGet, "/api/catalog/packages/{package_id}/versions/{version}", handler.resolvePackageVersion)
}

func (handler *Handler) resolvePackageVersion(writer http.ResponseWriter, request *http.Request) {
	id, err := pathPackageID(request, "package_id")
	var version string
	if err == nil {
		version, err = pathVersion(request)
	}
	if err == nil {
		err = requireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.package_version.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetPackageVersion(request.Context(), access, id, version)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
