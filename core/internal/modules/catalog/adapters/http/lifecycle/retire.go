package lifecyclehttp

import "net/http"

func (handler *Handler) RetireRoute() Route {
	return route(http.MethodPost, "/api/catalog/installations/{installation_id}/retirements", handler.retire)
}

func (handler *Handler) retire(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "installation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err == nil {
		err = requireNoBody(request)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRetire, "catalog.installation.retire", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.Retire(request.Context(), access, id, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
