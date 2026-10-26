package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identitySystemResourceCommand struct {
	ctx                context.Context
	scope              platformtx.TenantScope
	evidence           identityPublicEvidence
	resource           publicresource.Resource
	permissionID       uuid.UUID
	readPermissionID   uuid.UUID
	catalog            publicresource.FieldCatalog
	authorityOperation string
}

func (port *identitySystemResourcePort) beginSystemResourceCommand(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	resource publicresource.Resource,
	authorityOperation string,
	action string,
) (identitySystemResourceCommand, error) {
	evidence, err := newIdentityPublicEvidence(actor)
	exactScope := identitySystemResourceScope(
		ctx, scope, evidence, platformtx.Command, authorityOperation,
	) == nil
	delegatedScope := identitySystemResourceImportMutationScope(
		ctx, scope, evidence, resource, authorityOperation, action,
	)
	if err != nil || !resource.Valid() || !exactScope && !delegatedScope {
		return identitySystemResourceCommand{}, publicresource.ErrContract
	}
	authority, err := newIdentitySystemResourceAuthorization(
		port.application, ctx, scope, evidence, resource,
	)
	if err != nil {
		return identitySystemResourceCommand{}, err
	}
	permissionID, err := authority.require(action)
	if err != nil {
		return identitySystemResourceCommand{}, err
	}
	readPermissionID, err := authority.require("read")
	if err != nil {
		return identitySystemResourceCommand{}, err
	}
	catalog, err := port.resolveCatalog(ctx, scope, resource, authority)
	if err != nil {
		return identitySystemResourceCommand{}, err
	}
	return identitySystemResourceCommand{
		ctx: ctx, scope: scope, evidence: evidence, resource: resource,
		permissionID: permissionID, readPermissionID: readPermissionID,
		catalog: catalog, authorityOperation: authorityOperation,
	}, nil
}

func (command identitySystemResourceCommand) operation(
	idempotencyKey string,
) uuid.UUID {
	return identityOperationID(
		command.evidence, command.authorityOperation, idempotencyKey,
	)
}
