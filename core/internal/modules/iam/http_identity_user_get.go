package iam

import "net/http"

func (handler *IdentityHTTP) getUser(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	id, idValid := identityPathID(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	if !idValid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.GetUser(r.Context(), claims, id)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityUserView(view) || view.ID != id {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
