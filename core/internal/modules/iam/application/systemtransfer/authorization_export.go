package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ExportScope interface {
	TenantID() uuid.UUID
	ActorID() uuid.UUID
	IAMScope() platformtx.TenantScope
}

type FreshExportScopes interface {
	InFreshExportCheckpoint(
		context.Context,
		ExportAttempt,
		func(context.Context, ExportScope) error,
	) error
}

type ExportCheckpointAuthorizer interface {
	AuthorizeExportCheckpoint(
		context.Context,
		platformtx.TenantScope,
		ExportAttempt,
	) (publicresource.ActorEvidence, error)
}
