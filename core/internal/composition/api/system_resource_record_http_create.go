package api

import (
	"net/http"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) create(
	resource publicresource.Resource,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission, err := handler.recordAdmission(
			r, resource, systemResourceRecordHTTPCreate,
		)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		key, keyValid := systemResourceRecordIdempotencyKey(r)
		if !keyValid || r.URL.RawQuery != "" {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPRequestInvalid)
			return
		}
		document, err := decodeSystemResourceRecordDocument(w, r)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		defer destroySystemResourceRecordDocument(document)
		prepared, err := handler.mintRecordCreate(
			r.Context(), admission, resource, document, key,
		)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		defer prepared.Destroy()
		result, err := handler.application.Create(r.Context(), admission, prepared)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		handler.writeRecordHTTPMutation(w, r, http.StatusCreated, result, resource)
	})
}
