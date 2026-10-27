package lifecyclehttp

import "net/http"

func (handler *Handler) InstallDemoRoute() Route {
	return route(http.MethodPost,
		"/api/catalog/installations/{installation_id}/demo", handler.installDemo)
}

func (handler *Handler) installDemo(writer http.ResponseWriter, request *http.Request) {
	installationID, err := pathV7(request, "installation_id")
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
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(
		request, PermissionDemoManage, "catalog.installation.demo.install", "",
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	if handler.demo == nil {
		handler.failure(writer, request, ErrUnavailable)
		return
	}
	value, err := handler.demo.InstallDemo(request.Context(), access, installationID, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
