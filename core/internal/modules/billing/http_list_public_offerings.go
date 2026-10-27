package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listPublicOfferings(w http.ResponseWriter, r *http.Request) {
	request, err := handler.decodeGETQuery(r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.public.ListPublicOfferings(
		r.Context(), billingapi.PublicQuery{Request: request, Transport: http.MethodGet},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
