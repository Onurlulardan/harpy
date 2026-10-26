package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) platformQuery(
	w http.ResponseWriter,
	r *http.Request,
	permission billingapi.Permission,
	operation string,
	post bool,
) (billingapi.Query, error) {
	access, err := requireAccess(handler, r, permission, operation)
	if err != nil {
		return billingapi.Query{}, err
	}
	request, err := handler.decodeGETQuery(r)
	if post {
		request, err = handler.decodePOSTQuery(w, r)
	}
	if err != nil {
		return billingapi.Query{}, err
	}
	transport := http.MethodGet
	if post {
		transport = http.MethodPost
	}
	return billingapi.Query{Access: access, Request: request, Transport: transport}, nil
}
