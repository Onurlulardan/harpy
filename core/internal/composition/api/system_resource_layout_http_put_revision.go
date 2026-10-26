package api

import "net/http"

func (handler *systemResourceLayoutHTTPHandler) putRevision(w http.ResponseWriter, r *http.Request) {
	resource, resourceValid := layoutHTTPResource(r)
	revisionID, revisionValid := layoutHTTPRevisionID(r)
	if !resourceValid {
		handler.writeHTTPFailure(w, r, errLayoutHTTPResourceUnknown)
		return
	}
	admission, err := handler.admissionFor(
		r, resource, layoutHTTPActionUpdate, "system_resource_layout.create",
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	if !revisionValid || !layoutHTTPQueryAbsent(r) {
		handler.writeHTTPFailure(w, r, errLayoutHTTPRequestInvalid)
		return
	}
	definition, err := decodeLayoutHTTPDefinition(w, r)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	revision, err := handler.application.Create(
		r.Context(), admission, resource, revisionID, definition,
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	handler.writeHTTPRevision(w, r, http.StatusOK, resource, revision,
		layoutHTTPRevisionLocation(resource, revision.ID))
}
