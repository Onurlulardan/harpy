package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func identitySystemResourceScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	mode platformtx.ScopeMode,
	operation string,
) error {
	if ctx == nil || scope.TenantID() != evidence.Tenant() ||
		scope.ActorID() != evidence.Subject() || scope.TenantSlug() == "" ||
		scope.Mode() != mode || scope.OperationID() != operation {
		return publicresource.ErrContract
	}
	return nil
}

func identitySystemResourceLayoutPublishScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
) error {
	if !identitySystemResourceLayoutPublishIntent(scope.Mode(), scope.OperationID()) ||
		identitySystemResourceScope(
			ctx, scope, evidence, platformtx.Command,
			publicresource.LayoutPublishCatalogOperation,
		) != nil ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		!identitySystemResourceLayoutPublishOwners(scope) {
		return publicresource.ErrContract
	}
	return nil
}

func identitySystemResourceLayoutPublishOwners(
	scope platformtx.TenantScope,
) bool {
	iam, iamErr := platformtx.MintOwnerAuthority(scope, "iam")
	layout, layoutErr := platformtx.MintOwnerAuthority(scope, "system_resource_layout")
	return iamErr == nil && layoutErr == nil &&
		iam.OwnsScope(scope) && layout.OwnsScope(scope) &&
		iam.GroupID() == publicresource.AuthorityGroup &&
		layout.GroupID() == publicresource.AuthorityGroup &&
		iam.GroupGeneration() == iam.Generation() &&
		layout.GroupGeneration() == layout.Generation() &&
		iam.GroupGeneration() == layout.GroupGeneration()
}

func identitySystemResourceLayoutPublishIntent(
	mode platformtx.ScopeMode,
	operation string,
) bool {
	return mode == platformtx.Command &&
		operation == publicresource.LayoutPublishCatalogOperation
}
