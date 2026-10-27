package iam

import "net/http"

func (handler *RBACHTTP) getAction(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	id, idValid := rbacPathID(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	if !idValid || r.URL.RawQuery != "" {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.GetAction(r.Context(), claims, id)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
