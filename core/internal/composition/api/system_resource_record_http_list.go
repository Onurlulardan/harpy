package api

import (
	"net/http"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) list(
	resource systemresources.Resource,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission, err := handler.recordAdmission(
			r, resource, systemResourceRecordHTTPRead,
		)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		if !systemResourceRecordBodyAbsent(r) {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPRequestInvalid)
			return
		}
		request, err := iammodule.DecodeSystemResourceGETQuery(r.URL.RawQuery)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		input, err := systemresources.NewQueryInput(resource, request)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, systemresources.ErrContract)
			return
		}
		result, err := handler.application.Query(r.Context(), admission, input)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		handler.writeRecordHTTPQuery(w, r, result)
	})
}
