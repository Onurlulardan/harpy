package iam

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type identityRoleCreateDocument struct {
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	ScopeKind        string          `json:"scope_kind"`
	ScopeID          string          `json:"scope_id"`
	OrganizationID   string          `json:"organization_id"`
	Default          *bool           `json:"is_default"`
	CustomAttributes json.RawMessage `json:"custom_attributes"`
}

func (handler *IdentityHTTP) createRole(w http.ResponseWriter, r *http.Request) {
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
	document, custom, err := decodeIdentityProfileDocument[identityRoleCreateDocument](
		w, r, []string{"name", "description", "scope_kind", "scope_id",
			"organization_id", "is_default"},
	)
	if err != nil {
		handler.writeIdentityInputFailure(w, r, err)
		return
	}
	document.CustomAttributes = custom
	input, valid := identityRoleCreateInput(key, claims.Tenant(), document)
	if !valid {
		handler.writeIdentityFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.CreateRole(r.Context(), claims, input)
	if err != nil {
		handler.writeIdentityApplicationFailure(w, r, err)
		return
	}
	if !validIdentityRoleView(view) {
		handler.report(ErrIdentityHTTPUnavailable)
		handler.writeIdentityFailure(w, r, identityFailureUnavailable)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}

func identityRoleCreateInput(
	key string, tenantID uuid.UUID, document identityRoleCreateDocument,
) (IdentityRoleCreate, bool) {
	organizationID := uuid.Nil
	organizationValid := document.OrganizationID == ""
	if document.OrganizationID != "" {
		organizationID, organizationValid = decodeUUID(document.OrganizationID)
	}
	scopeKind, scopeID, scopeValid := identityRoleScopeInput(
		tenantID, organizationID, document.ScopeKind, document.ScopeID,
	)
	validScope := scopeKind == "TENANT" && organizationID == uuid.Nil ||
		scopeKind == "ORGANIZATION" && organizationID == scopeID
	valid := scopeValid && organizationValid && validScope && document.Default != nil &&
		validIdentityText(document.Name, 200) && len(document.Description) <= 2000
	return IdentityRoleCreate{
		IdempotencyKey: key, Name: document.Name, Description: document.Description,
		ScopeKind: scopeKind, ScopeID: scopeID,
		OrganizationID: organizationID, Default: document.Default != nil && *document.Default,
		CustomAttributes: identityCustomAttributes(document.CustomAttributes),
	}, valid
}

func identityRoleScopeInput(
	tenantID uuid.UUID,
	organizationID uuid.UUID,
	scopeKind string,
	rawScopeID string,
) (string, uuid.UUID, bool) {
	if scopeKind == "" && rawScopeID == "" {
		if organizationID != uuid.Nil {
			return "ORGANIZATION", organizationID, true
		}
		return "TENANT", tenantID, tenantID != uuid.Nil
	}
	if scopeKind == "" || rawScopeID == "" {
		return "", uuid.Nil, false
	}
	scopeID, valid := decodeUUID(rawScopeID)
	return scopeKind, scopeID, valid
}
