package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) CreateRole(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityRoleCreate,
) (IdentityRoleView, error) {
	profile, err := identityProfileValues(input.CustomAttributes)
	if err != nil || !identityRoleScopeValid(claims, input) {
		return IdentityRoleView{}, ErrIdentityHTTPInvalid
	}
	operationID := identityOperationID(
		claims, "iam.identity.role.create", input.IdempotencyKey,
	)
	roleID := identityChildID(operationID, "role")
	requestSHA, err := identityRequestSHA(struct {
		Name, Description, ScopeKind string
		ScopeID, OrganizationID      uuid.UUID
		Default                      bool
		Profile                      []byte
	}{
		input.Name, input.Description, input.ScopeKind, input.ScopeID,
		input.OrganizationID, input.Default, profile,
	})
	if err != nil {
		return IdentityRoleView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.role.create", IdentityRoles, "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			storedProfile, profileErr := application.profileDocumentValues(
				scoped, scope, identityfacade.ProfileRole, profile,
			)
			if profileErr != nil {
				return profileErr
			}
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, createErr := application.identities.roles.CreateRole(
				scoped, scope, roleapp.CreateRoleCommand{
					ID:             identityfacade.IdentityID(roleID),
					ScopeKind:      identityfacade.RoleScopeKind(strings.ToLower(input.ScopeKind)),
					ScopeID:        identityfacade.IdentityID(input.ScopeID),
					OrganizationID: identityfacade.IdentityID(input.OrganizationID),
					Name:           input.Name, Description: input.Description, Default: input.Default,
					ProfileValues: storedProfile,
					Evidence:      evidence,
				},
			)
			if createErr != nil {
				return createErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, roleID, "role",
				"iam.identity.role.create", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityRoleView{}, err
	}
	return application.getRole(ctx, claims, roleID, "create")
}

func identityRoleScopeValid(claims accesstoken.Claims, input IdentityRoleCreate) bool {
	return input.IdempotencyKey != "" && input.Name != "" &&
		(input.ScopeKind == "TENANT" && input.ScopeID == claims.Tenant() &&
			input.OrganizationID == uuid.Nil ||
			input.ScopeKind == "ORGANIZATION" && input.ScopeID != uuid.Nil &&
				input.ScopeID == input.OrganizationID)
}
