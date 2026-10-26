package iam

import "net/http"

func (handler *PortalUserHTTP) invite(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	id, key, inputValid := portalUserPathAndKey(r)
	if !valid {
		handler.writeFailure(w, r, portalUserFailureAuthentication)
		return
	}
	if !inputValid {
		handler.writeFailure(w, r, portalUserFailureInvalid)
		return
	}
	if r.ContentLength > 0 {
		if _, err := decodeIdentityDocument[portalUserEmptyDocument](w, r); err != nil {
			handler.writeInputFailure(w, r, err)
			return
		}
	}
	view, err := handler.application.InvitePortalUser(r.Context(), claims,
		PortalUserMutationInput{ID: id, IdempotencyKey: key})
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeView(w, r, http.StatusAccepted, view)
}
