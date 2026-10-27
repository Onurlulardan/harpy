package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func identitySystemResourceQueryInput(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.QueryInput,
) (identityPublicEvidence, publicresource.Resource, publicresource.QueryRequest, error) {
	evidence, evidenceErr := newIdentityPublicEvidence(actor)
	resource, resourceErr := input.Resource()
	request, requestErr := input.Request()
	if evidenceErr != nil || resourceErr != nil || requestErr != nil ||
		identitySystemResourceScope(
			ctx, scope, evidence, platformtx.SnapshotQuery, "restquery.execute",
		) != nil {
		return identityPublicEvidence{}, "", nil,
			publicresource.ErrContract
	}
	return evidence, resource, request, nil
}
