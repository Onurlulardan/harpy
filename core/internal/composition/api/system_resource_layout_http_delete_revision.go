package api

import (
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

func (handler *systemResourceLayoutHTTPHandler) deleteRevision(w http.ResponseWriter, r *http.Request) {
	resource, resourceValid := layoutHTTPResource(r)
	revisionID, revisionValid := layoutHTTPRevisionID(r)
	if !resourceValid {
		handler.writeHTTPFailure(w, r, errLayoutHTTPResourceUnknown)
		return
	}
	admission, err := handler.admissionFor(
		r, resource, layoutHTTPActionUpdate, "system_resource_layout.delete",
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	expected, etagValid, missing := layoutHTTPExpectedRevision(r)
	if missing {
		handler.writeHTTPFailure(w, r, errLayoutHTTPPrecondition)
		return
	}
	if !revisionValid || !etagValid || !layoutHTTPQueryAbsent(r) ||
		!layoutHTTPBodyAbsent(r) {
		handler.writeHTTPFailure(w, r, errLayoutHTTPRequestInvalid)
		return
	}
	err = handler.application.Delete(
		r.Context(), admission, resource,
		layouts.DeleteInput{ID: revisionID, ExpectedRowRevision: expected},
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	handler.writeHTTPDelete(w, r, revisionID)
}
