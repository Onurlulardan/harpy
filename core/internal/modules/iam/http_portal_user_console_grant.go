package iam

import "net/http"

func (handler *PortalUserHTTP) grantConsole(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	input, inputValid := portalUserMutation(r)
	if !valid {
		handler.writeFailure(w, r, portalUserFailureAuthentication)
		return
	}
	if !inputValid {
		handler.writeFailure(w, r, portalUserFailurePrecondition)
		return
	}
	if r.ContentLength > 0 {
		if _, err := decodeIdentityDocument[portalUserEmptyDocument](w, r); err != nil {
			handler.writeInputFailure(w, r, err)
			return
		}
	}
	view, err := handler.application.GrantPortalConsole(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeView(w, r, http.StatusOK, view)
}
