package api

import "net/http"

func (handler *systemResourceLayoutHTTPHandler) getRevision(w http.ResponseWriter, r *http.Request) {
	resource, resourceValid := layoutHTTPResource(r)
	revisionID, revisionValid := layoutHTTPRevisionID(r)
	if !resourceValid {
		handler.writeHTTPFailure(w, r, errLayoutHTTPResourceUnknown)
		return
	}
	admission, err := handler.admissionFor(
		r, resource, layoutHTTPActionRead, "system_resource_layout.get",
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	if !revisionValid || !layoutHTTPQueryAbsent(r) || !layoutHTTPBodyAbsent(r) {
		handler.writeHTTPFailure(w, r, errLayoutHTTPRequestInvalid)
		return
	}
	revision, err := handler.application.Get(
		r.Context(), admission, resource, revisionID,
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	handler.writeHTTPRevision(w, r, http.StatusOK, resource, revision, "")
}
