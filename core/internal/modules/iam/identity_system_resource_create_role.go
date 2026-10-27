package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) createSystemRole(
	command identitySystemResourceCommand,
	claim identitySystemResourceCreateClaim,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	name, _, nameErr := values.text(publicresource.RoleNameFieldID, true)
	description, _, descriptionErr := values.text(publicresource.RoleDescriptionFieldID, false)
	scopeKind, _, scopeErr := values.text(publicresource.RoleScopeKindFieldID, true)
	scopeID, scopeFound, scopeIDErr := values.nullableUUID(publicresource.RoleScopeIDFieldID)
	organizationID, organizationFound, organizationErr := values.nullableUUID(
		publicresource.RoleOrganizationIDFieldID,
	)
	defaultValue, _, defaultErr := values.boolean(publicresource.RoleDefaultFieldID)
	if nameErr != nil || descriptionErr != nil || scopeErr != nil || scopeIDErr != nil ||
		organizationErr != nil || defaultErr != nil || !scopeFound {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	kind := identityfacade.RoleScopeKind(scopeKind)
	if kind == identityfacade.RoleScopeTenant {
		if scopeID != command.scope.TenantID() || organizationFound && organizationID != uuid.Nil {
			return publicresource.MutationResult{}, publicresource.ErrContract
		}
		organizationID = uuid.Nil
	} else if kind != identityfacade.RoleScopeOrganization ||
		!organizationFound || organizationID == uuid.Nil || scopeID != organizationID {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	profile, err := port.createNameProfile(
		command, identityfacade.ProfileRole, values, name,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	operationID := command.operation(claim.idempotencyKey)
	requestSHA, err := identitySystemResourceCreateDigest(
		claim.resource, claim.values, [32]byte{},
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	roleID := identityChildID(operationID, "role")
	result, err := port.application.identities.roles.CreateRole(
		command.ctx, command.scope, roleapp.CreateRoleCommand{
			ID: identityfacade.IdentityID(roleID), ScopeKind: kind,
			ScopeID:        identityfacade.IdentityID(scopeID),
			OrganizationID: identityfacade.IdentityID(organizationID),
			Name:           name, Description: description, Default: defaultValue,
			ProfileValues: profile, Evidence: evidence,
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, roleID, "role",
		command.authorityOperation, result.Replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, roleID, operationID, requestSHA, result.Revision, result.Replayed,
	)
}
