package formrecordarchive

import (
	"context"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type AccessRefresher interface {
	Refresh(context.Context, foundation.ID, foundation.ID, uint64, []foundation.ID) error
}

type GenerationResolver interface {
	ResolveTenantBillingGenerations(context.Context, uuid.UUID) (int64, int64, int64, error)
	ResolveTenantDocumentsGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type ScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type ArchiveOwner interface {
	Create(context.Context, platformtx.TenantScope,
		archiveapp.CreateRequest) (archiveapp.Snapshot, error)
	Get(context.Context, platformtx.TenantScope,
		archiveapp.GetRequest) (archiveapp.Snapshot, error)
}

type LinkMatcher interface {
	Match(context.Context, platformtx.TenantScope, domain.ID, domain.ID,
		domain.ID, domain.ID, []domain.ID) error
}

type Dependencies struct {
	Access      AccessRefresher
	Generations GenerationResolver
	Scopes      ScopeCoordinator
	Archives    ArchiveOwner
	Links       LinkMatcher
}

type Application struct{ dependencies Dependencies }

func NewApplication(dependencies Dependencies) (*Application, error) {
	for _, dependency := range []any{dependencies.Access, dependencies.Generations,
		dependencies.Scopes, dependencies.Archives, dependencies.Links} {
		if foundation.NilPort(dependency) {
			return nil, formrecordarchiveapi.ErrUnavailable
		}
	}
	return &Application{dependencies: dependencies}, nil
}

var _ formrecordarchiveapi.Owner = (*Application)(nil)
