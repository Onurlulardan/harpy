package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type systemTransferBatchScopes struct {
	generations IdentityHTTPPlatform
	scopes      SystemTransferScopeCoordinator
}

type systemTransferBatchScope struct {
	tenantID, actorID uuid.UUID
	scope             platformtx.TenantScope
}

func newSystemTransferBatchScopes(generations IdentityHTTPPlatform,
	scopes SystemTransferScopeCoordinator) (*systemTransferBatchScopes, error) {
	if nilIdentityDependency(generations) || nilIdentityDependency(scopes) {
		return nil, systemtransfer.ErrDependencies
	}
	return &systemTransferBatchScopes{generations: generations, scopes: scopes}, nil
}

func (runtime *systemTransferBatchScopes) InFreshImportBatch(
	ctx context.Context,
	attempt systemtransfer.Attempt,
	execute func(context.Context, systemtransfer.BatchScope) error,
) error {
	if runtime == nil || ctx == nil || ctx.Err() != nil || attempt.Validate() != nil ||
		execute == nil {
		return systemtransfer.ErrInvalidAttempt
	}
	values, err := runtime.generations.ResolveSnapshotDomainGenerations(
		ctx, attempt.TenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(values) {
		return systemtransfer.ErrAuthorizationCheckpoint
	}
	request, err := platformtx.AdmitCommandScope(
		attempt.ActorID, attempt.TenantID, SystemImportExecuteOperation,
		identityAuthorityGroup, values["iam"], values,
	)
	if err != nil {
		return systemtransfer.ErrAuthorizationCheckpoint
	}
	return runtime.scopes.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			return execute(scoped, systemTransferBatchScope{
				tenantID: attempt.TenantID, actorID: attempt.ActorID, scope: scope,
			})
		})
}

func (scope systemTransferBatchScope) TenantID() uuid.UUID { return scope.tenantID }
func (scope systemTransferBatchScope) ActorID() uuid.UUID  { return scope.actorID }
func (scope systemTransferBatchScope) IAMScope() platformtx.TenantScope {
	return scope.scope
}

var _ systemtransfer.FreshBatchScopes = (*systemTransferBatchScopes)(nil)
