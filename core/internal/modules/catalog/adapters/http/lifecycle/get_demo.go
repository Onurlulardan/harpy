package lifecyclehttp

import "net/http"

func (handler *Handler) GetDemoStatusRoute() Route {
	return route(http.MethodGet,
		"/api/catalog/installations/{installation_id}/demo", handler.getDemoStatus)
}

func (handler *Handler) getDemoStatus(writer http.ResponseWriter, request *http.Request) {
	installationID, err := pathV7(request, "installation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err == nil {
		err = requireNoBody(request)
	}
	if err != nil {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(
		request, PermissionRead, "catalog.installation.demo.get", "",
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	if handler.demo == nil {
		handler.failure(writer, request, ErrUnavailable)
		return
	}
	value, err := handler.demo.GetDemoStatus(request.Context(), access, installationID)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
