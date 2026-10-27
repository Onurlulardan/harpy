package registryhttp

import "net/http"

func (handler *Handler) ListPackagesRoute() Route {
	return route(http.MethodGet, "/api/catalog/packages", handler.listPackages)
}

func (handler *Handler) listPackages(writer http.ResponseWriter, request *http.Request) {
	limit, after, err := packagePageInput(request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.package.list", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	items, next, err := handler.app.ListPackages(request.Context(), access, limit, after)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, struct {
		Items []Package `json:"items"`
		Next  any       `json:"next"`
	}{Items: items, Next: next})
}
