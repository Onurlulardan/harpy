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
