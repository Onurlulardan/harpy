package lifecyclehttp

import "net/http"

type removeDemoDocument struct {
	ExpectedPlanSHA256 string `json:"expected_plan_sha256"`
}

func (handler *Handler) RemoveDemoRoute() Route {
	return route(http.MethodDelete,
		"/api/catalog/installations/{installation_id}/demo", handler.removeDemo)
}

func (handler *Handler) removeDemo(writer http.ResponseWriter, request *http.Request) {
	installationID, err := pathV7(request, "installation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	var document removeDemoDocument
	if err == nil {
		err = decodeBody(request, &document)
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
		request, PermissionDemoManage, "catalog.installation.demo.remove", "",
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	if handler.demo == nil {
		handler.failure(writer, request, ErrUnavailable)
		return
	}
	value, err := handler.demo.RemoveDemo(request.Context(), access,
		installationID, document.ExpectedPlanSHA256, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
