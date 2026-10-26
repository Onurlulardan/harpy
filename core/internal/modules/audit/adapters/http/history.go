package audithttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

func (handler *Handler) entityHistory(response http.ResponseWriter, request *http.Request) {
	handler.history(response, request, application.OperationEntityHistory, func(filter *domain.Filter) error {
		filter.EntityType, filter.EntityID = request.PathValue("entity_type"),
			request.PathValue("entity_id")
		return filter.Validate()
	})
}

func (handler *Handler) formRecordHistory(response http.ResponseWriter, request *http.Request) {
	handler.history(response, request, application.OperationFormRecordHistory, func(filter *domain.Filter) error {
		filter.FormSlug, filter.EntityID = request.PathValue("form_slug"),
			request.PathValue("record_id")
		return filter.Validate()
	})
}

func (handler *Handler) history(
	response http.ResponseWriter, request *http.Request, operation string,
	bind func(*domain.Filter) error,
) {
	access, err := handler.accessFor(request, operation)
	var filter domain.Filter
	if err == nil {
		filter, err = historyFilter(request)
	}
	if err == nil && bind(&filter) != nil {
		err = application.ErrInvalid
	}
	if err == nil {
		var page domain.Page
		page, err = handler.dependencies.Application.List(request.Context(), access, filter, operation)
		if err == nil {
			data := pageView(page)
			data["total_changes"] = page.Total
			if operation == application.OperationEntityHistory {
				data["entity_type"], data["entity_id"] = filter.EntityType, filter.EntityID
			} else {
				data["form_slug"], data["record_id"] = filter.FormSlug, filter.EntityID
				data["timeline"] = data["changes"]
			}
			handler.writeSuccess(response, request, data)
			return
		}
	}
	handler.writeFailure(response, request, err)
}
