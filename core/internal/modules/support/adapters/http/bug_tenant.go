package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (handler *Handler) listTenantBugs(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeTenant,
		application.PermissionBugRead, application.OperationBugTenantList)
	filter := domain.BugFilter{}
	if err == nil {
		filter, err = bugFilter(request, 20, false)
	}
	if err == nil {
		var page domain.BugPage
		page, err = handler.applications.Tenant.ListTenantBugs(request.Context(), access, filter)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugPageView(page))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) tenantBugStats(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeTenant,
		application.PermissionBugRead, application.OperationBugTenantStats)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	if err == nil {
		var stats domain.BugStats
		stats, err = handler.applications.Tenant.ReadTenantBugStats(request.Context(), access)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugStatsView(stats))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) getTenantBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeTenant,
		application.PermissionBugRead, application.OperationBugTenantRead)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Tenant.GetTenantBug(request.Context(), access, domain.ID(id))
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) updateTenantBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeTenant,
		application.PermissionBugManage, application.OperationBugTenantUpdate)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	var update domain.BugUpdate
	if err == nil {
		update, err = decodeBugUpdate(response, request, domain.ID(id), access.TenantID())
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Tenant.UpdateTenantBug(request.Context(), access, update)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}
