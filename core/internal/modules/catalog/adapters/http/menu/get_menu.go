package menuhttp

import "net/http"

func (handler *Handler) GetMenuRoute() Route {
	return route(http.MethodGet, "/api/catalog/menu", handler.getMenu)
}

func (handler *Handler) getMenu(writer http.ResponseWriter, request *http.Request) {
	if err := requireNoQuery(request); err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.menu.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetMenu(request.Context(), access)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
