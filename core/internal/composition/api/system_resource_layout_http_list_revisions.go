package api

import "net/http"

func (handler *systemResourceLayoutHTTPHandler) listRevisions(w http.ResponseWriter, r *http.Request) {
	resource, valid := layoutHTTPResource(r)
	if !valid {
		handler.writeHTTPFailure(w, r, errLayoutHTTPResourceUnknown)
		return
	}
	admission, err := handler.admissionFor(
		r, resource, layoutHTTPActionRead, "system_resource_layout.list",
	)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	if !layoutHTTPQueryAbsent(r) || !layoutHTTPBodyAbsent(r) {
		handler.writeHTTPFailure(w, r, errLayoutHTTPRequestInvalid)
		return
	}
	revisions, err := handler.application.List(r.Context(), admission, resource)
	if err != nil {
		handler.writeHTTPFailure(w, r, err)
		return
	}
	handler.writeHTTPRevisionList(w, r, resource, revisions)
}
