package lifecyclehttp

import "net/http"

func (handler *Handler) GetInstallationRoute() Route {
	return route(http.MethodGet, "/api/catalog/installations/{installation_id}", handler.getInstallation)
}

func (handler *Handler) getInstallation(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "installation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.installation.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetInstallation(request.Context(), access, id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
