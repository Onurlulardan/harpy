package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// ResourceMutator is the complete IAM mutation surface available to imports.
// Query, delete and state transitions intentionally remain unreachable.
type ResourceMutator interface {
	PrepareCreate(context.Context, publicresource.CreateInput) (
		publicresource.PreparedCreate, error,
	)
	Create(
		context.Context,
		platformtx.TenantScope,
		publicresource.ActorEvidence,
		publicresource.PreparedCreate,
	) (publicresource.MutationResult, error)
	Update(
		context.Context,
		platformtx.TenantScope,
		publicresource.ActorEvidence,
		publicresource.UpdateInput,
	) (publicresource.MutationResult, error)
}
