package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func (port *identitySystemResourcePort) deleteSystemResource(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
) (publicresource.DeleteResult, error) {
	operationID := command.operation(deletion.idempotencyKey)
	requestSHA, err := identitySystemResourceTerminalDigest(
		command.resource, deletion.id, deletion.expectedRevision, "delete",
	)
	if err != nil {
		return publicresource.DeleteResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	var revision int64
	var replayed bool
	switch command.resource {
	case publicresource.Users:
		revision, replayed, err = port.deleteSystemResourceUser(
			command, deletion, operationID, requestSHA, evidence,
		)
	case publicresource.Roles:
		result, runErr := port.application.identities.roles.RetireRole(
			command.ctx, command.scope, roleapp.RetireRoleCommand{
				ID:               identityfacade.IdentityID(deletion.id),
				ExpectedRevision: deletion.expectedRevision, Evidence: evidence,
			},
		)
		revision, replayed, err = result.Revision, result.Replayed, runErr
	case publicresource.Organizations:
		result, runErr := port.application.identities.organizations.DeleteOrganization(
			command.ctx, command.scope, organizationapp.OrganizationStateCommand{
				ID:               identityfacade.IdentityID(deletion.id),
				ExpectedRevision: deletion.expectedRevision, Evidence: evidence,
			},
		)
		revision, replayed, err = result.Revision, result.Replayed, runErr
	default:
		return publicresource.DeleteResult{}, publicresource.ErrContract
	}
	if err != nil {
		return publicresource.DeleteResult{}, identitySystemResourceError(err)
	}
	return port.finishSystemResourceDelete(
		command, deletion.id, operationID, requestSHA, revision, replayed, evidence,
	)
}
