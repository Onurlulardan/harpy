package iam

import (
	"encoding/json"
	"net/http"
)

type identityUserCreateDocument struct {
	Email            string          `json:"email"`
	Password         json.RawMessage `json:"password"`
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Phone            string          `json:"phone"`
	OrganizationID   string          `json:"organization_id"`
	RoleIDs          []string        `json:"role_ids"`
	Status           string          `json:"status"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) createUser(w http.ResponseWriter, r *http.Request) {
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
	document, custom, err := decodeIdentityProfileDocument[identityUserCreateDocument](
		w, r, []string{"email", "password", "first_name", "last_name", "phone",
			"organization_id", "role_ids", "status"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	password, err := decodeLoginPassword(document.Password)
	clear(document.Password)
	organizationID, organizationValid := decodeUUID(document.OrganizationID)
	roleIDs, rolesValid := decodeUUIDList(document.RoleIDs)
	if err != nil || !organizationValid || !rolesValid || document.Status != "ACTIVE" ||
		!validIdentityText(document.Email, 254) ||
		!validIdentityText(document.FirstName, 200) ||
		!validIdentityText(document.LastName, 200) ||
		!validIdentityOptionalText(document.Phone, 32) {
		clear(password)
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	input := IdentityUserCreate{
		IdempotencyKey: key, Email: document.Email, Password: password,
		FirstName: document.FirstName, LastName: document.LastName,
		Phone:          document.Phone,
		OrganizationID: organizationID, RoleIDs: roleIDs,
		CustomAttributes: identityCustomAttributes(custom),
	}
	defer input.Destroy()
	view, err := handler.application.CreateUser(r.Context(), claims, input)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityUserView(view) {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}
