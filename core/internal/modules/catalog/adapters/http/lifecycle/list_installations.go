package lifecyclehttp

import "net/http"

func (handler *Handler) ListInstallationsRoute() Route {
	return route(http.MethodGet, "/api/catalog/installations", handler.listInstallations)
}

func (handler *Handler) listInstallations(writer http.ResponseWriter, request *http.Request) {
	limit, after, err := pageInput(request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.installation.list", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	items, next, err := handler.app.ListInstallations(request.Context(), access, limit, after)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, struct {
		Items []Installation `json:"items"`
		Next  any            `json:"next"`
	}{Items: items, Next: next})
}
