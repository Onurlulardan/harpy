package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identitySystemResourceState struct {
	id               uuid.UUID
	expectedRevision int64
	state            publicresource.State
	idempotencyKey   string
}

func (port *identitySystemResourcePort) ChangeState(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.StateInput,
) (publicresource.MutationResult, error) {
	resource, resourceErr := input.Resource()
	id, expectedRevision, identityErr := input.Identity()
	state, stateErr := input.State()
	key, keyErr := input.IdempotencyKey()
	operation, action := identitySystemResourceStateAuthority(resource)
	if resourceErr != nil || identityErr != nil || stateErr != nil || keyErr != nil ||
		operation == "" {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	command, err := port.beginSystemResourceCommand(
		ctx, scope, actor, resource, operation, action,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	change := identitySystemResourceState{
		id: uuid.UUID(id), expectedRevision: expectedRevision,
		state: state, idempotencyKey: key,
	}
	return port.changeSystemResourceState(command, change)
}

func identitySystemResourceStateAuthority(
	resource publicresource.Resource,
) (string, string) {
	switch resource {
	case publicresource.Users:
		return "iam.identity.user.disable", "update"
	case publicresource.Roles:
		return "iam.identity.role.retire", "delete"
	case publicresource.Organizations:
		return "iam.identity.organization.disable", "update"
	default:
		return "", ""
	}
}
