package iam

import "net/http"

type identityRoleStateDocument struct {
	State string `json:"state"`
}

func (handler *IdentityHTTP) changeRoleState(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	mutation, valid := identityMutationFromRequest(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailurePrecondition)
		return
	}
	document, err := decodeIdentityDocument[identityRoleStateDocument](w, r)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	if document.State != "RETIRED" {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.ChangeRoleState(r.Context(), claims,
		IdentityStateMutation{IdentityMutation: mutation, State: document.State})
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityRoleView(view) || view.ID != mutation.ID {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
