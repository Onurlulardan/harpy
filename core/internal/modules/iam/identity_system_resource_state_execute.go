package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
)

func (port *identitySystemResourcePort) changeSystemResourceState(
	command identitySystemResourceCommand,
	change identitySystemResourceState,
) (publicresource.MutationResult, error) {
	operationID := command.operation(change.idempotencyKey)
	requestSHA, err := identitySystemResourceTerminalDigest(
		command.resource, change.id, change.expectedRevision, string(change.state),
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	var revision int64
	var replayed bool
	switch command.resource {
	case publicresource.Users:
		result, runErr := port.application.identities.users.DisableUser(
			command.ctx, command.scope, userapp.UserStateCommand{
				ID:               identityfacade.IdentityID(change.id),
				ExpectedRevision: change.expectedRevision, Evidence: evidence,
			},
		)
		revision, replayed, err = result.Revision, result.Replayed, runErr
	case publicresource.Roles:
		result, runErr := port.application.identities.roles.RetireRole(
			command.ctx, command.scope, roleapp.RetireRoleCommand{
				ID:               identityfacade.IdentityID(change.id),
				ExpectedRevision: change.expectedRevision, Evidence: evidence,
			},
		)
		revision, replayed, err = result.Revision, result.Replayed, runErr
	case publicresource.Organizations:
		result, runErr := port.application.identities.organizations.DisableOrganization(
			command.ctx, command.scope, organizationapp.OrganizationStateCommand{
				ID:               identityfacade.IdentityID(change.id),
				ExpectedRevision: change.expectedRevision, Evidence: evidence,
			},
		)
		revision, replayed, err = result.Revision, result.Replayed, runErr
	default:
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.finishSystemResourceState(
		command, change.id, operationID, requestSHA, revision, replayed, evidence,
	)
}
