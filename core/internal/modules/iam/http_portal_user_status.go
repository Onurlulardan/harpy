package iam

import "net/http"

func (handler *PortalUserHTTP) changeStatus(w http.ResponseWriter, r *http.Request) {
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
	document, err := decodeIdentityDocument[portalUserStatusDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	input.Status = normalizePortalStatus(document.Status)
	if input.Status != "ACTIVE" && input.Status != "PENDING" && input.Status != "INACTIVE" {
		handler.writeFailure(w, r, portalUserFailureInvalid)
		return
	}
	view, err := handler.application.ChangePortalUserStatus(r.Context(), claims, input)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeView(w, r, http.StatusOK, view)
}
