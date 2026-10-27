package iam

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type identityOrganizationCreateDocument struct {
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	ParentID         string          `json:"parent_id"`
	OwnerID          json.RawMessage `json:"owner_id"`
	OwnerIDs         json.RawMessage `json:"owner_ids"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) createOrganization(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureAuthentication)
		return
	}
	key, valid := identityIdempotencyKey(r)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	document, custom, err := decodeIdentityProfileDocument[identityOrganizationCreateDocument](
		w, r, []string{"name", "slug", "parent_id", "owner_id", "owner_ids"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	parentID, parentValid := optionalCreateParent(document.ParentID)
	ownerIDs, _, ownersValid := decodeIdentityOrganizationOwners(
		document.OwnerID, document.OwnerIDs, true,
	)
	if !parentValid || !validIdentityText(document.Name, 200) ||
		!validIdentityText(document.Slug, 100) || !ownersValid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	input := IdentityOrganizationCreate{
		IdempotencyKey: key, Name: document.Name, Slug: document.Slug,
		ParentID: parentID, OwnerIDs: ownerIDs,
		CustomAttributes: identityCustomAttributes(custom),
	}
	view, err := handler.application.CreateOrganization(r.Context(), claims, input)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityOrganizationView(view) {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}

func optionalCreateParent(value string) (uuid.UUID, bool) {
	if value == "" {
		return uuid.Nil, true
	}
	return decodeUUID(value)
}
