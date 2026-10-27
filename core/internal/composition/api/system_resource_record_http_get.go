package api

import (
	"net/http"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) get(
	resource systemresources.Resource,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission, err := handler.recordReadAdmission(r, resource)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		id, idValid := systemResourceRecordID(r)
		if !idValid || r.URL.RawQuery != "" || !systemResourceRecordBodyAbsent(r) {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPRequestInvalid)
			return
		}
		input, err := systemresources.NewGetInput(resource, id)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, systemresources.ErrContract)
			return
		}
		result, err := handler.application.Get(r.Context(), admission, input)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		handler.writeRecordHTTPRead(w, r, http.StatusOK, result, "")
	})
}
