package iam

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferExportScopes struct {
	generations IdentityHTTPPlatform
	scopes      SystemTransferScopeCoordinator
}

func newSystemTransferExportScopes(
	generations IdentityHTTPPlatform,
	scopes SystemTransferScopeCoordinator,
) (*systemTransferExportScopes, error) {
	if nilIdentityDependency(generations) || nilIdentityDependency(scopes) {
		return nil, systemtransfer.ErrExportUnavailable
	}
	return &systemTransferExportScopes{generations: generations, scopes: scopes}, nil
}

func (runtime *systemTransferExportScopes) generationsFor(
	ctx context.Context,
	attempt systemtransfer.ExportAttempt,
) (map[string]int64, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || attempt.Validate() != nil {
		return nil, systemtransfer.ErrExportAuthorizationRevoked
	}
	values, err := runtime.generations.ResolveSnapshotDomainGenerations(
		ctx, attempt.TenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(values) {
		return nil, systemtransfer.ErrExportAuthorizationRevoked
	}
	return values, nil
}

func (runtime *systemTransferExportScopes) InFreshExportCheckpoint(
	ctx context.Context,
	attempt systemtransfer.ExportAttempt,
	execute func(context.Context, systemtransfer.ExportScope) error,
) error {
	if execute == nil {
		return systemtransfer.ErrExportWorkerInvalid
	}
	values, err := runtime.generationsFor(ctx, attempt)
	if err != nil {
		return err
	}
	request, err := platformtx.AdmitCommandScope(
		attempt.ActorID, attempt.TenantID, SystemExportExecuteOperation,
		identityAuthorityGroup, values["iam"], values,
	)
	if err != nil {
		return systemtransfer.ErrExportAuthorizationRevoked
	}
	return runtime.scopes.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			return execute(scoped, systemTransferBatchScope{tenantID: attempt.TenantID,
				actorID: attempt.ActorID, scope: scope})
		})
}

func (runtime *systemTransferExportScopes) withSnapshot(
	ctx context.Context,
	attempt systemtransfer.ExportAttempt,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	if execute == nil {
		return systemtransfer.ErrExportWorkerInvalid
	}
	values, err := runtime.generationsFor(ctx, attempt)
	if err != nil {
		return err
	}
	request, err := queryapp.SystemResourceSnapshotScopeRequest(
		attempt.ActorID, attempt.TenantID, values,
	)
	if err != nil {
		return systemtransfer.ErrExportAuthorizationRevoked
	}
	return runtime.scopes.InTenant(ctx, request, execute)
}

var _ systemtransfer.FreshExportScopes = (*systemTransferExportScopes)(nil)
