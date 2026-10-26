package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (handler *Handler) listPlatformBugs(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformBugRead, application.OperationBugPlatformList)
	filter := domain.BugFilter{}
	if err == nil {
		filter, err = bugFilter(request, 20, true)
	}
	if err == nil {
		var page domain.BugPage
		page, err = handler.applications.Platform.ListPlatformBugs(request.Context(), access, filter)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugPageView(page))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) platformBugStats(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformBugRead, application.OperationBugPlatformStats)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	if err == nil {
		var stats domain.BugStats
		stats, err = handler.applications.Platform.ReadPlatformBugStats(request.Context(), access)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugStatsView(stats))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) getPlatformBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformBugRead, application.OperationBugPlatformRead)
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	tenantID := uuid.Nil
	if err == nil && !onlyQueryKeys(request.URL.Query(), "tenant_id") {
		err = application.ErrInvalid
	}
	if err == nil {
		tenantID, err = exactUUID(request.URL.Query().Get("tenant_id"))
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Platform.GetPlatformBug(request.Context(), access,
			domain.ID(id), domain.ID(tenantID))
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) updatePlatformBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformBugManage, application.OperationBugPlatformUpdate)
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	tenantID := uuid.Nil
	if err == nil && !onlyQueryKeys(request.URL.Query(), "tenant_id") {
		err = application.ErrInvalid
	}
	if err == nil {
		tenantID, err = exactUUID(request.URL.Query().Get("tenant_id"))
	}
	var update domain.BugUpdate
	if err == nil {
		update, err = decodeBugUpdate(response, request, domain.ID(id), domain.ID(tenantID))
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Platform.UpdatePlatformBug(request.Context(), access, update)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}
