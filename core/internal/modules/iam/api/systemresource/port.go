package systemresource

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Port interface {
	Catalog(context.Context, transaction.TenantScope, ActorEvidence, Resource) (FieldCatalog, error)
	PrepareCreate(context.Context, CreateInput) (PreparedCreate, error)
	Create(context.Context, transaction.TenantScope, ActorEvidence, PreparedCreate) (MutationResult, error)
	Get(context.Context, transaction.TenantScope, ActorEvidence, GetInput) (ReadResult, error)
	Query(context.Context, transaction.TenantScope, ActorEvidence, QueryInput) (QueryResult, error)
	Update(context.Context, transaction.TenantScope, ActorEvidence, UpdateInput) (MutationResult, error)
	Delete(context.Context, transaction.TenantScope, ActorEvidence, DeleteInput) (DeleteResult, error)
	ChangeState(context.Context, transaction.TenantScope, ActorEvidence, StateInput) (MutationResult, error)
}
