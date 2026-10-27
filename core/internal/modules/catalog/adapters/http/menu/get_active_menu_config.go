package menuhttp

import "net/http"

func (handler *Handler) GetActiveMenuConfigRoute() Route {
	return route(http.MethodGet, "/api/form-engine/menu/configs/active", handler.getActiveMenuConfig)
}

func (handler *Handler) getActiveMenuConfig(writer http.ResponseWriter, request *http.Request) {
	if err := requireNoQuery(request); err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.menu.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetActiveMenuConfig(request.Context(), access)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
