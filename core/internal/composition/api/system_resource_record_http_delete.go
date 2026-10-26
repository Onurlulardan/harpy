package api

import (
	"net/http"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) delete(
	resource publicresource.Resource,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admission, err := handler.recordAdmission(
			r, resource, systemResourceRecordHTTPDelete,
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
		if !idValid || !revisionValid || !keyValid || r.URL.RawQuery != "" ||
			!systemResourceRecordBodyAbsent(r) {
			handler.writeRecordHTTPFailure(w, r, errSystemResourceRecordHTTPRequestInvalid)
			return
		}
		input, err := mintRecordDelete(resource, id, revision, key)
		if err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		if err = handler.application.Delete(r.Context(), admission, input); err != nil {
			handler.writeRecordHTTPFailure(w, r, err)
			return
		}
		handler.writeRecordHTTPDelete(w, r)
	})
}
