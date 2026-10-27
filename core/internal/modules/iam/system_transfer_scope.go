package iam

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferGenerations struct {
	iam    int64
	values map[string]int64
}

func (runtime *SystemTransferRuntime) resolveSystemTransfer(
	ctx context.Context,
	access SystemTransferAccess,
	resource SystemTransferResource,
	operation string,
) (systemTransferGenerations, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !resource.Valid() ||
		!access.ValidFor(resource, resource.Permission(systemTransferAction(operation)), operation) {
		return systemTransferGenerations{}, ErrSystemTransferForbidden
	}
	values, err := runtime.generations.ResolveSnapshotDomainGenerations(
		ctx, access.TenantID(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(values) {
		return systemTransferGenerations{}, ErrSystemTransferUnavailable
	}
	return systemTransferGenerations{iam: values["iam"], values: values}, nil
}

func (runtime *SystemTransferRuntime) systemTransferCommand(
	ctx context.Context,
	access SystemTransferAccess,
	values systemTransferGenerations,
	operation string,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := platformtx.AdmitCommandScope(
		access.PrincipalID(), access.TenantID(), operation,
		identityAuthorityGroup, values.iam, values.values,
	)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	return systemTransferError(runtime.scopes.InTenant(ctx, request, execute))
}

func (runtime *SystemTransferRuntime) systemTransferSnapshot(
	ctx context.Context,
	access SystemTransferAccess,
	values systemTransferGenerations,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := queryapp.SnapshotScopeRequest(
		access.PrincipalID(), access.TenantID(), map[string]int64{"iam": values.iam},
	)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	return systemTransferError(runtime.scopes.InTenant(ctx, request, execute))
}
