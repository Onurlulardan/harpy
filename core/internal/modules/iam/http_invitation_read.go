package iam

import "net/http"

func (handler *InvitationHTTP) read(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	id, idValid := identityPathID(r)
	if !valid {
		handler.writeFailure(w, r, errInvitationAuthentication)
		return
	}
	if !idValid {
		handler.writeFailure(w, r, ErrInvitationHTTPInvalid)
		return
	}
	view, err := handler.application.Read(r.Context(), claims, id)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeInvitation(w, r, http.StatusOK, view)
}
