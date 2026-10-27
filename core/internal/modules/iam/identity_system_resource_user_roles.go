package iam

import (
	"context"
	"strings"

	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) userRoleGrantsInScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	permissionID uuid.UUID,
	organizationID uuid.UUID,
	roleIDs []uuid.UUID,
) ([]identityUserRoleGrant, error) {
	return application.userRoleGrantsForOrganizationsInScope(
		ctx, scope, evidence, permissionID, []uuid.UUID{organizationID}, true, roleIDs,
	)
}

func (application *identityHTTPApplication) userRoleGrantsForOrganizationsInScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	permissionID uuid.UUID,
	organizationIDs []uuid.UUID,
	requireOrganization bool,
	roleIDs []uuid.UUID,
) ([]identityUserRoleGrant, error) {
	ordered := identitySortedIDs(roleIDs)
	organizations := make(map[uuid.UUID]struct{}, len(organizationIDs))
	for _, organizationID := range organizationIDs {
		organizations[organizationID] = struct{}{}
	}
	grants := make([]identityUserRoleGrant, 0, len(ordered))
	for _, roleID := range ordered {
		role, err := application.identities.roles.GetRole(
			ctx, scope, roleapp.GetRoleQuery{
				ID:       identityfacade.IdentityID(roleID),
				Evidence: identityQueryEvidence(evidence, permissionID),
			},
		)
		if err != nil {
			return nil, err
		}
		_, organizationFound := organizations[uuid.UUID(role.OrganizationID)]
		if strings.ToLower(string(role.Status)) != "active" ||
			requireOrganization && role.ScopeKind == identityfacade.RoleScopeOrganization &&
				!organizationFound {
			return nil, ErrIdentityHTTPInvalid
		}
		grants = append(grants, identityUserRoleGrant{
			roleID: roleID, scopeKind: role.ScopeKind,
			scopeID:        uuid.UUID(role.ScopeID),
			organizationID: uuid.UUID(role.OrganizationID),
		})
	}
	return grants, nil
}
