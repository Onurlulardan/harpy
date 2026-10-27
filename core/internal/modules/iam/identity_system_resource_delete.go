package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identitySystemResourceDelete struct {
	id               uuid.UUID
	expectedRevision int64
	idempotencyKey   string
}

func (port *identitySystemResourcePort) Delete(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.DeleteInput,
) (publicresource.DeleteResult, error) {
	resource, resourceErr := input.Resource()
	id, expectedRevision, identityErr := input.Identity()
	key, keyErr := input.IdempotencyKey()
	operation, action := identitySystemResourceDeleteAuthority(resource)
	if resourceErr != nil || identityErr != nil || keyErr != nil || operation == "" {
		return publicresource.DeleteResult{}, publicresource.ErrContract
	}
	command, err := port.beginSystemResourceCommand(
		ctx, scope, actor, resource, operation, action,
	)
	if err != nil {
		return publicresource.DeleteResult{}, err
	}
	deletion := identitySystemResourceDelete{
		id: uuid.UUID(id), expectedRevision: expectedRevision, idempotencyKey: key,
	}
	return port.deleteSystemResource(command, deletion)
}

func identitySystemResourceDeleteAuthority(
	resource publicresource.Resource,
) (string, string) {
	switch resource {
	case publicresource.Users:
		return "iam.identity.user.delete", "delete"
	case publicresource.Roles:
		return "iam.identity.role.retire", "delete"
	case publicresource.Organizations:
		return "iam.identity.organization.delete", "delete"
	default:
		return "", ""
	}
}
