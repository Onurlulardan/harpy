package audithttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

func (handler *Handler) list(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.OperationList)
	var filter domain.Filter
	if err == nil {
		filter, err = listFilter(request)
	}
	if err == nil {
		var page domain.Page
		page, err = handler.dependencies.Application.List(
			request.Context(), access, filter, application.OperationList)
		if err == nil {
			handler.writeSuccess(response, request, pageView(page))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) stats(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.OperationStats)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	if err == nil {
		var stats domain.Stats
		stats, err = handler.dependencies.Application.Stats(request.Context(), access)
		if err == nil {
			handler.writeSuccess(response, request, statsView(stats))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) get(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.OperationGet)
	if err == nil && len(request.URL.Query()) != 0 {
		err = application.ErrInvalid
	}
	var id uuid.UUID
	if err == nil {
		id, err = exactEntryUUID(request.PathValue("id"))
	}
	if err == nil {
		entry, getErr := handler.dependencies.Application.Get(request.Context(), access, id)
		if getErr == nil {
			view := entryView(entry)
			handler.writeSuccess(response, request,
				map[string]any{"change": view, "field_changes": view["field_changes"]})
			return
		}
		err = getErr
	}
	handler.writeFailure(response, request, err)
}
