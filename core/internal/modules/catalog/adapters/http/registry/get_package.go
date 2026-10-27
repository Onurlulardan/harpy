package registryhttp

import "net/http"

func (handler *Handler) GetPackageRoute() Route {
	return route(http.MethodGet, "/api/catalog/packages/{package_id}", handler.getPackage)
}

func (handler *Handler) getPackage(writer http.ResponseWriter, request *http.Request) {
	id, err := pathPackageID(request, "package_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.package.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetPackage(request.Context(), access, id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
