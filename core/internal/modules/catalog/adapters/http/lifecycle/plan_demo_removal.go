package lifecyclehttp

import "net/http"

func (handler *Handler) PlanDemoRemovalRoute() Route {
	return route(http.MethodGet,
		"/api/catalog/installations/{installation_id}/demo/removal-plan",
		handler.planDemoRemoval)
}

func (handler *Handler) planDemoRemoval(writer http.ResponseWriter, request *http.Request) {
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
		request, PermissionDemoManage, "catalog.installation.demo.removal_plan", "",
	)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	if handler.demo == nil {
		handler.failure(writer, request, ErrUnavailable)
		return
	}
	value, err := handler.demo.PlanDemoRemoval(request.Context(), access, installationID)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
