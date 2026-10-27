package portaluser

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Store interface {
	Create(context.Context, platformtx.TenantScope, CreateCommand) (MutationResult, error)
	ChangeStatus(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	GrantConsole(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	RevokeConsole(context.Context, platformtx.TenantScope, RevisionCommand) (MutationResult, error)
	Read(context.Context, platformtx.TenantScope, uuid.UUID) (View, error)
	ProjectSubjectLinks(
		context.Context, platformtx.TenantScope, ProjectionAccess,
		uuid.UUID, uuid.UUID, []uuid.UUID,
	) ([]SubjectLink, error)
	VerifyAccess(
		context.Context, platformtx.TenantScope, uuid.UUID, uuid.UUID, int64, int64,
	) (AccessBinding, error)
}
