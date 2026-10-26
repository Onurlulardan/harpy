package api

import (
	"net/http"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) update(
	resource publicresource.Resource,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission, err := handler.recordAdmission(
			r, resource, systemResourceRecordHTTPUpdate,
		)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		id, idValid := systemResourceRecordID(r)
		revision, revisionValid, revisionMissing := systemResourceRecordExpectedRevision(r)
		key, keyValid := systemResourceRecordIdempotencyKey(r)
		if revisionMissing {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPPrecondition)
			return
		}
		if !idValid || !revisionValid || !keyValid || r.URL.RawQuery != "" {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPRequestInvalid)
			return
		}
		document, err := decodeSystemResourceRecordDocument(w, r)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		defer destroySystemResourceRecordDocument(document)
		input, err := handler.mintRecordUpdate(
			r.Context(), admission, resource, id, revision, document, key,
		)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		result, err := handler.application.Update(r.Context(), admission, input)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		handler.writeRecordHTTPMutation(w, r, http.StatusOK, result, resource)
	})
}
