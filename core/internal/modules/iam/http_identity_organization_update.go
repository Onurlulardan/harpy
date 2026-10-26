package iam

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type identityOrganizationUpdateDocument struct {
	Name             *string         `json:"name"`
	Slug             *string         `json:"slug"`
	ParentID         *string         `json:"parent_id"`
	OwnerID          json.RawMessage `json:"owner_id"`
	OwnerIDs         json.RawMessage `json:"owner_ids"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) updateOrganization(w http.ResponseWriter, r *http.Request) {
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
	document, custom, err := decodeIdentityProfileDocument[identityOrganizationUpdateDocument](
		w, r, []string{"name", "slug", "parent_id", "owner_id", "owner_ids"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	document.CustomAttributes = custom
	parentID, parentValid := decodeOptionalUUID(document.ParentID)
	ownerIDs, ownersFound, ownersValid := decodeIdentityOrganizationOwners(
		document.OwnerID, document.OwnerIDs, false,
	)
	fields := document.Name != nil || document.Slug != nil ||
		document.ParentID != nil || ownersFound || len(document.CustomAttributes) != 0
	if !fields || !parentValid || !ownersValid || parentID != nil && *parentID == mutation.ID ||
		document.Name != nil && !validIdentityText(*document.Name, 200) ||
		document.Slug != nil && !validIdentityText(*document.Slug, 100) {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	var ownerSet *[]uuid.UUID
	if ownersFound {
		ownerSet = &ownerIDs
	}
	input := IdentityOrganizationUpdate{
		IdentityMutation: mutation, Name: document.Name, Slug: document.Slug,
		ParentID: parentID, OwnerIDs: ownerSet,
		CustomAttributes: document.CustomAttributes,
	}
	view, err := handler.application.UpdateOrganization(r.Context(), claims, input)
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
