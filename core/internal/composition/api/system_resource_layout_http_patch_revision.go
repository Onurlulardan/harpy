package api

import (
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

func (handler *systemResourceLayoutHTTPHandler) patchRevision(w http.ResponseWriter, r *http.Request) {
	resource, resourceValid := layoutHTTPResource(r)
	revisionID, revisionValid := layoutHTTPRevisionID(r)
	if !resourceValid {
		handler.writeHTTPFailure(w, r, errLayoutHTTPResourceUnknown)
		return
	}
	admission, err := handler.admissionFor(
		r, resource, layoutHTTPActionUpdate, "system_resource_layout.update",
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
	if !revisionValid || !etagValid || !layoutHTTPQueryAbsent(r) {
		handler.writeHTTPFailure(w, r, errLayoutHTTPRequestInvalid)
		return
	}
	definition, err := decodeLayoutHTTPDefinition(w, r)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	revision, err := handler.application.Update(
		r.Context(), admission, resource, layouts.UpdateInput{
			ID: revisionID, ExpectedRowRevision: expected, Definition: definition,
		},
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	handler.writeHTTPRevision(w, r, http.StatusOK, resource, revision, "")
}
