package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type BatchScope interface {
	TenantID() uuid.UUID
	ActorID() uuid.UUID
	IAMScope() platformtx.TenantScope
}

// BatchAuthorizer performs the fresh IAM permission check inside the command scope.
type BatchAuthorizer interface {
	AuthorizeImportBatch(
		context.Context,
		platformtx.TenantScope,
		Attempt,
	) (publicresource.ActorEvidence, error)
}

type FreshBatchScopes interface {
	InFreshImportBatch(
		context.Context,
		Attempt,
		func(context.Context, BatchScope) error,
	) error
}
