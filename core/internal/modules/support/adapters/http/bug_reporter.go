package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (handler *Handler) createBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeReporter,
		application.PermissionBugCreate, application.OperationBugCreate)
	var document bugCreateDocument
	var canonical []byte
	if err == nil {
		canonical, err = decodeObject(response, request, &document)
	}
	var key string
	var digest [32]byte
	if err == nil {
		key, digest, err = commandEvidence(request, canonical)
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Reporter.CreateBug(request.Context(), access,
			bugInput(document), key, digest)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusCreated, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) listMyBugs(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeReporter,
		application.PermissionBugReadOwn, application.OperationBugMine)
	filter := domain.BugFilter{}
	if err == nil {
		filter, err = bugFilter(request, 10, false)
	}
	if err == nil {
		var page domain.BugPage
		page, err = handler.applications.Reporter.ListMyBugs(request.Context(), access, filter)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugPageView(page))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) getMyBug(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopeReporter,
		application.PermissionBugReadOwn, application.OperationBugRead)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	id, idErr := pathUUID(request)
	if err == nil {
		err = idErr
	}
	if err == nil {
		var report domain.BugReport
		report, err = handler.applications.Reporter.GetMyBug(request.Context(), access, domain.ID(id))
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, bugView(report))
			return
		}
	}
	handler.writeFailure(response, request, err)
}
