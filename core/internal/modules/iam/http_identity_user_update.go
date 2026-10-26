package iam

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type identityUserUpdateDocument struct {
	FirstName        *string         `json:"first_name"`
	LastName         *string         `json:"last_name"`
	Phone            *string         `json:"phone"`
	OrganizationID   *string         `json:"organization_id"`
	RoleIDs          *[]string       `json:"role_ids"`
	Status           string          `json:"status"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) updateUser(w http.ResponseWriter, r *http.Request) {
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
	document, custom, err := decodeIdentityProfileDocument[identityUserUpdateDocument](
		w, r, []string{"first_name", "last_name", "phone", "organization_id",
			"role_ids", "status"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	document.CustomAttributes = custom
	input, valid := identityUserUpdateInput(mutation, document)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.UpdateUser(r.Context(), claims, input)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityUserView(view) || view.ID != mutation.ID {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusOK, view.Revision, view)
}

func identityUserUpdateInput(
	mutation IdentityMutation,
	document identityUserUpdateDocument,
) (IdentityUserUpdate, bool) {
	organizationID, organizationValid := decodeOptionalUUID(document.OrganizationID)
	var roleIDs *[]uuid.UUID
	rolesValid := true
	if document.RoleIDs != nil {
		decoded, valid := decodeUUIDList(*document.RoleIDs)
		rolesValid = valid
		roleIDs = &decoded
	}
	fields := document.FirstName != nil || document.LastName != nil || document.Phone != nil ||
		document.OrganizationID != nil || document.RoleIDs != nil ||
		document.Status != "" || len(document.CustomAttributes) != 0
	valid := fields && organizationValid && rolesValid &&
		(document.FirstName == nil || validIdentityText(*document.FirstName, 200)) &&
		(document.LastName == nil || validIdentityText(*document.LastName, 200)) &&
		(document.Phone == nil || validIdentityOptionalText(*document.Phone, 32)) &&
		(document.Status == "" || document.Status == "INACTIVE")
	return IdentityUserUpdate{
		IdentityMutation: mutation, FirstName: document.FirstName,
		LastName: document.LastName, OrganizationID: organizationID,
		Phone:   document.Phone,
		RoleIDs: roleIDs, CustomAttributes: document.CustomAttributes,
		LegacyState: document.Status,
	}, valid
}
