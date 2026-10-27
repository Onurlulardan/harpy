package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *identitySystemResourcePort) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.PreparedCreate,
) (publicresource.MutationResult, error) {
	claim, err := claimIdentitySystemResourceCreate(input)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	defer func() { claim.credential = identityPreparedUserCredential{} }()
	operation := identitySystemResourceCreateOperation(claim.resource)
	command, err := port.beginSystemResourceCommand(
		ctx, scope, actor, claim.resource, operation, "create",
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	values, err := newIdentitySystemResourceValues(
		claim.resource, claim.values, command.catalog,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	switch claim.resource {
	case publicresource.Users:
		return port.createSystemUser(command, claim, values)
	case publicresource.Roles:
		return port.createSystemRole(command, claim, values)
	case publicresource.Organizations:
		return port.createSystemOrganization(command, claim, values)
	default:
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
}

func identitySystemResourceCreateOperation(resource publicresource.Resource) string {
	switch resource {
	case publicresource.Users:
		return "iam.identity.user.create"
	case publicresource.Roles:
		return "iam.identity.role.create"
	case publicresource.Organizations:
		return "iam.identity.organization.create"
	default:
		return ""
	}
}
