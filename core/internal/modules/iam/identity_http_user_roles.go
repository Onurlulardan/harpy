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

type identityUserRoleGrant struct {
	roleID         uuid.UUID
	scopeKind      identityfacade.RoleScopeKind
	scopeID        uuid.UUID
	organizationID uuid.UUID
}

func (application *identityHTTPApplication) userRoleGrants(
	ctx context.Context,
	claims accesstoken.Claims,
	organizationID uuid.UUID,
	roleIDs []uuid.UUID,
) ([]identityUserRoleGrant, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	ordered := identitySortedIDs(roleIDs)
	grants := make([]identityUserRoleGrant, 0, len(ordered))
	err := application.inSnapshotWithAction(
		ctx, claims, "iam.identity.role.read", IdentityUsers, "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			for _, roleID := range ordered {
				role, err := application.identities.roles.GetRole(
					scoped, scope, roleapp.GetRoleQuery{
						ID:       identityfacade.IdentityID(roleID),
						Evidence: identityQueryEvidence(claims, permissionID),
					},
				)
				if err != nil {
					return err
				}
				if strings.ToLower(string(role.Status)) != "active" ||
					role.ScopeKind == identityfacade.RoleScopeOrganization &&
						uuid.UUID(role.OrganizationID) != organizationID {
					return ErrIdentityHTTPInvalid
				}
				grants = append(grants, identityUserRoleGrant{
					roleID: roleID, scopeKind: role.ScopeKind,
					scopeID:        uuid.UUID(role.ScopeID),
					organizationID: uuid.UUID(role.OrganizationID),
				})
			}
			return nil
		},
	)
	return grants, err
}
