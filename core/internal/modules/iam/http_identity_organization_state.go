package iam

import "net/http"

type identityOrganizationStateDocument struct {
	State string `json:"state"`
}

func (handler *IdentityHTTP) changeOrganizationState(w http.ResponseWriter, r *http.Request) {
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
	document, err := decodeIdentityDocument[identityOrganizationStateDocument](w, r)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	if document.State != "SUSPENDED" {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.ChangeOrganizationState(r.Context(), claims,
		IdentityStateMutation{IdentityMutation: mutation, State: document.State})
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityOrganizationView(view) || view.ID != mutation.ID {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}
