package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userrelationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/userrelation"
)

func (port *identitySystemResourcePort) updateSystemUser(
	command identitySystemResourceCommand,
	update identitySystemResourceUpdate,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	if _, found := values.raw(publicresource.UserEmailFieldID); found {
		return publicresource.MutationResult{}, publicresource.ErrForbidden
	}
	organizations, replaceOrganizations, err := values.uuidList(
		publicresource.UserOrganizationIDsFieldID, false,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	roles, replaceRoles, err := values.uuidList(publicresource.UserRoleIDsFieldID, false)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	profile, err := values.profile()
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	patchProfile := string(profile) != "{}"
	if !patchProfile && !replaceOrganizations && !replaceRoles {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	if (replaceOrganizations || replaceRoles) && update.id == command.evidence.Subject() {
		return publicresource.MutationResult{}, publicresource.ErrForbidden
	}
	if replaceRoles {
		if _, err = port.application.userRoleGrantsForOrganizationsInScope(
			command.ctx, command.scope, command.evidence, command.permissionID,
			organizations, replaceOrganizations, roles,
		); err != nil {
			return publicresource.MutationResult{}, identitySystemResourceError(err)
		}
	}
	operationID := command.operation(update.idempotencyKey)
	requestSHA, err := identitySystemResourceUpdateDigest(
		command.resource, update.id, update.expectedRevision, update.values,
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	organizationPatch, err := userrelationapp.NewOrganizationSetPatch(
		identityfacade.IdentityID(update.id), replaceOrganizations,
		identityFacadeIDs(organizations),
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	rolePatch, err := userrelationapp.NewRoleSetPatch(
		identityfacade.IdentityID(update.id), replaceRoles, identityFacadeIDs(roles),
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if !patchProfile {
		profile = nil
	}
	result, err := port.application.identities.userRelations.Replace(
		command.ctx, command.scope, userrelationapp.ReplaceCommand{
			UserID: identityfacade.IdentityID(update.id), ExpectedRevision: update.expectedRevision,
			CommandKey:   userrelationapp.UpdateCommandKey,
			PatchProfile: patchProfile, ProfilePatch: profile,
			Organizations: organizationPatch, Roles: rolePatch, Evidence: evidence,
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, update.id, "user",
		command.authorityOperation, result.Replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, update.id, operationID, requestSHA,
		result.Revision, result.Replayed,
	)
}
