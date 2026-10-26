package iam

import (
	"context"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	userrelationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/userrelation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func identityUserUpdateRequestSHA(
	input IdentityUserUpdate,
	custom []byte,
) ([32]byte, error) {
	var roleIDs *[]uuid.UUID
	if input.RoleIDs != nil {
		ordered := identitySortedIDs(*input.RoleIDs)
		roleIDs = &ordered
	}
	return identityRequestSHA(struct {
		ID               uuid.UUID
		ExpectedRevision int64
		FirstName        *string
		LastName         *string
		Phone            *string
		OrganizationID   *uuid.UUID
		RoleIDs          *[]uuid.UUID
		Custom           []byte
		State            string
	}{
		input.ID, input.ExpectedRevision, input.FirstName, input.LastName, input.Phone,
		input.OrganizationID, roleIDs, custom, input.LegacyState,
	})
}

func (application *identityHTTPApplication) replaceLegacyUserRelations(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims identityEvidenceSource,
	permissionID uuid.UUID,
	input IdentityUserUpdate,
	custom []byte,
	profileChanged bool,
	evidence identityfacade.MutationEvidence,
) (identityfacade.MutationResult, error) {
	user, err := application.identities.users.GetUser(
		ctx, scope, userapp.GetUserQuery{
			ID:       identityfacade.IdentityID(input.ID),
			Evidence: identityQueryEvidence(claims, permissionID),
		},
	)
	if err != nil {
		return identityfacade.MutationResult{}, err
	}
	organizationIDs, roleIDs := identityLegacyUserRelationSets(user, input)
	if _, err = application.userRoleGrantsForOrganizationsInScope(
		ctx, scope, identityEvidence(claims), permissionID,
		organizationIDs, true, roleIDs,
	); err != nil {
		return identityfacade.MutationResult{}, err
	}
	organizationPatch, err := userrelationapp.NewOrganizationSetPatch(
		identityfacade.IdentityID(input.ID), input.OrganizationID != nil,
		identityReplacementIDs(input.OrganizationID != nil, organizationIDs),
	)
	if err != nil {
		return identityfacade.MutationResult{}, err
	}
	rolePatch, err := userrelationapp.NewRoleSetPatch(
		identityfacade.IdentityID(input.ID), input.RoleIDs != nil,
		identityReplacementIDs(input.RoleIDs != nil, roleIDs),
	)
	if err != nil {
		return identityfacade.MutationResult{}, err
	}
	var profile []byte
	if profileChanged {
		profile, err = application.userProfileUpdateValues(
			ctx, scope, user, custom,
			input.FirstName, input.LastName, input.Phone,
		)
		if err != nil {
			return identityfacade.MutationResult{}, err
		}
	}
	return application.identities.userRelations.Replace(
		ctx, scope, userrelationapp.ReplaceCommand{
			UserID: identityfacade.IdentityID(input.ID), ExpectedRevision: input.ExpectedRevision,
			CommandKey: userrelationapp.UpdateCommandKey, PatchProfile: profileChanged,
			ProfilePatch: profile, Organizations: organizationPatch,
			Roles: rolePatch, Evidence: evidence,
		},
	)
}

func identityLegacyUserRelationSets(
	user identityfacade.User,
	input IdentityUserUpdate,
) ([]uuid.UUID, []uuid.UUID) {
	organizationIDs := identityUUIDs(user.OrganizationIDs)
	if input.OrganizationID != nil {
		organizationIDs = []uuid.UUID{*input.OrganizationID}
	}
	roleIDs := identityUUIDs(user.RoleIDs)
	if input.RoleIDs != nil {
		roleIDs = append([]uuid.UUID(nil), (*input.RoleIDs)...)
	}
	return organizationIDs, roleIDs
}

func identityReplacementIDs(replace bool, values []uuid.UUID) []identityfacade.IdentityID {
	if !replace {
		return nil
	}
	return identityFacadeIDs(values)
}

func identityEvidence(source identityEvidenceSource) identityPublicEvidence {
	return identityPublicEvidence{
		tenantID: source.Tenant(), actorID: source.Subject(),
		tokenVersion: source.TokenVersion(), authorizationVersion: source.AuthorizationVersion(),
	}
}
