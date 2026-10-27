package iam

import (
	"encoding/json"
	"net/http"
)

type identityRoleUpdateDocument struct {
	Name             *string         `json:"name"`
	Description      *string         `json:"description"`
	Default          *bool           `json:"is_default"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) updateRole(w http.ResponseWriter, r *http.Request) {
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
	document, custom, err := decodeIdentityProfileDocument[identityRoleUpdateDocument](
		w, r, []string{"name", "description", "is_default"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	document.CustomAttributes = custom
	fields := document.Name != nil || document.Description != nil ||
		document.Default != nil || len(document.CustomAttributes) != 0
	if !fields || document.Name != nil && !validIdentityText(*document.Name, 200) ||
		document.Description != nil && len(*document.Description) > 2000 {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	input := IdentityRoleUpdate{
		IdentityMutation: mutation, Name: document.Name,
		Description: document.Description, Default: document.Default,
		CustomAttributes: document.CustomAttributes,
	}
	view, err := handler.application.UpdateRole(r.Context(), claims, input)
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
