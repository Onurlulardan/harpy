package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identitySystemResourceUpdate struct {
	id               uuid.UUID
	expectedRevision int64
	idempotencyKey   string
	values           publicresource.FieldValues
}

func (port *identitySystemResourcePort) Update(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.UpdateInput,
) (publicresource.MutationResult, error) {
	resource, resourceErr := input.Resource()
	id, expectedRevision, identityErr := input.Identity()
	values, valuesErr := input.Values()
	key, keyErr := input.IdempotencyKey()
	operation := identitySystemResourceUpdateOperation(resource)
	if resourceErr != nil || identityErr != nil || valuesErr != nil || keyErr != nil ||
		operation == "" {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	command, err := port.beginSystemResourceCommand(
		ctx, scope, actor, resource, operation, "update",
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	parsed, err := newIdentitySystemResourceValues(resource, values, command.catalog)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	update := identitySystemResourceUpdate{
		id: uuid.UUID(id), expectedRevision: expectedRevision,
		idempotencyKey: key, values: values,
	}
	switch resource {
	case publicresource.Users:
		return port.updateSystemUser(command, update, parsed)
	case publicresource.Roles:
		return port.updateSystemRole(command, update, parsed)
	case publicresource.Organizations:
		return port.updateSystemOrganization(command, update, parsed)
	default:
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
}

func identitySystemResourceUpdateOperation(resource publicresource.Resource) string {
	switch resource {
	case publicresource.Users:
		return "iam.identity.user.update"
	case publicresource.Roles:
		return "iam.identity.role.update"
	case publicresource.Organizations:
		return "iam.identity.organization.update"
	default:
		return ""
	}
}
