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
			data["entity_type"], data["entity_id"] = filter.EntityType, filter.EntityID
			handler.writeSuccess(response, request, data)
			return
		}
	}
	handler.writeFailure(response, request, err)
}
