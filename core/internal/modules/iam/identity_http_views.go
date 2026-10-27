package iam

import (
	"context"
	"encoding/json"
	"strings"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) userView(
	ctx context.Context,
	scope platformtx.TenantScope,
	user identityfacade.User,
) (IdentityUserView, error) {
	values, err := application.publicProfileValues(
		ctx, scope, identityfacade.ProfileUser, user.Profile,
	)
	if err != nil {
		return IdentityUserView{}, ErrIdentityHTTPUnavailable
	}
	first, firstValid := identityPublicString(values, "first_name", true)
	if !firstValid {
		first, firstValid = identityPublicString(values, "display_name", true)
	}
	last, lastValid := identityPublicString(values, "last_name", false)
	phone, phoneValid := identityPublicString(values, "phone", false)
	delete(values, "display_name")
	if !firstValid || !lastValid || !phoneValid {
		return IdentityUserView{}, ErrIdentityHTTPUnavailable
	}
	custom, err := identityMarshalProfile(values)
	if err != nil {
		return IdentityUserView{}, ErrIdentityHTTPUnavailable
	}
	organizationIDs := identityUUIDs(user.OrganizationIDs)
	return IdentityUserView{
		ID: uuid.UUID(user.ID), Email: user.EmailDisplay,
		FirstName: first, LastName: last, Status: identityUserStatus(user.Status),
		Phone: phone, EmailVerified: user.EmailVerified,
		AccountType: identityUserAccountType(user.AccessModes),
		RoleIDs:     identityUUIDs(user.RoleIDs), OrganizationIDs: organizationIDs,
		OrganizationID:   identitySingleOrganizationID(organizationIDs),
		CustomAttributes: custom, Revision: user.Revision,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func identitySingleOrganizationID(values []uuid.UUID) *uuid.UUID {
	if len(values) != 1 || values[0] == uuid.Nil {
		return nil
	}
	value := values[0]
	return &value
}

func identityPublicString(
	values map[string]json.RawMessage, key string, required bool,
) (string, bool) {
	raw, found := values[key]
	delete(values, key)
	if !found {
		return "", !required
	}
	var value string
	if json.Unmarshal(raw, &value) != nil || required && value == "" {
		return "", false
	}
	return value, true
}

func (application *identityHTTPApplication) roleView(
	ctx context.Context, scope platformtx.TenantScope, role identityfacade.Role,
) (IdentityRoleView, error) {
	custom, err := application.publicProfileDocument(
		ctx, scope, identityfacade.ProfileRole, role.Profile,
	)
	if err != nil {
		return IdentityRoleView{}, ErrIdentityHTTPUnavailable
	}
	return IdentityRoleView{
		ID: uuid.UUID(role.ID), Name: role.Name, Description: role.Description,
		ScopeKind: strings.ToUpper(string(role.ScopeKind)), ScopeID: uuid.UUID(role.ScopeID),
		OrganizationID: uuid.UUID(role.OrganizationID), Default: role.Default,
		Protected: role.Protected, Status: strings.ToUpper(string(role.Status)),
		CustomAttributes: custom, Revision: role.Revision,
		CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
	}, nil
}

func (application *identityHTTPApplication) organizationView(
	ctx context.Context, scope platformtx.TenantScope,
	organization identityfacade.Organization,
) (IdentityOrganizationView, error) {
	custom, err := application.publicProfileDocument(
		ctx, scope, identityfacade.ProfileOrganization, organization.Profile,
	)
	if err != nil {
		return IdentityOrganizationView{}, ErrIdentityHTTPUnavailable
	}
	ownerIDs := identityUUIDs(organization.OwnerIDs)
	return IdentityOrganizationView{
		ID: uuid.UUID(organization.ID), Name: organization.Name, Slug: organization.Slug,
		Status:   identityOrganizationStatus(organization.Status),
		ParentID: uuid.UUID(organization.ParentID), OwnerIDs: ownerIDs,
		OwnerID: identitySingleOrganizationID(ownerIDs), CustomAttributes: custom,
		Revision: organization.Revision, CreatedAt: organization.CreatedAt,
		UpdatedAt: organization.UpdatedAt,
	}, nil
}

func identityUserAccountType(modes []string) string {
	operator, portal := false, false
	for _, mode := range modes {
		operator = operator || mode == "operator_console"
		portal = portal || mode == "portal"
	}
	if operator && portal {
		return "operator_portal"
	}
	if operator {
		return "operator"
	}
	if portal {
		return "portal"
	}
	return ""
}

func identityUserStatus(status identityfacade.UserStatus) string {
	if status == identityfacade.UserTombstoned {
		return "DELETED"
	}
	if status == identityfacade.UserDisabled || status == identityfacade.UserSuspended {
		return "INACTIVE"
	}
	return "ACTIVE"
}

func identityOrganizationStatus(status identityfacade.OrganizationStatus) string {
	if status == identityfacade.OrganizationTombstoned {
		return "DELETED"
	}
	return strings.ToUpper(string(status))
}

func identityUUIDs(values []identityfacade.IdentityID) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	for index, value := range values {
		result[index] = uuid.UUID(value)
	}
	return result
}
