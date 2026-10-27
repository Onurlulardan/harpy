package work

import (
	"context"
	"fmt"

	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/boards"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/calendar"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/checklists"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/columns"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/comments"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/labels"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/memberships"
	projectionapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/projections"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/recurrence"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantWorkGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantBillingGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type PlatformHTTPRuntimeDependencies struct {
	Generations         HTTPGenerationResolver
	Scopes              HTTPScopeCoordinator
	Catalogs            platformasync.WriteCatalogs
	Principals          workprincipalapi.Directory
	CursorKeys          RESTQueryCursorKeys
	ProjectionResources ports.ResourceAuthority
	TaskReferences      ports.ResourceReferenceRegistry
	TaskResources       LegacyTaskResourceResolver
}

type httpServices struct {
	boards      *boards.Service
	columns     *columns.Service
	memberships *memberships.Service
	tasks       *tasks.Service
	comments    *comments.Service
	labels      *labels.Service
	checklists  *checklists.Service
	calendar    *calendar.Service
	recurrence  *recurrence.Service
}

type PlatformHTTPRuntime struct {
	resolver      HTTPGenerationResolver
	scopes        HTTPScopeCoordinator
	principals    workprincipalapi.Directory
	unit          *workpostgres.Binder
	reader        *workpostgres.Reader
	authorizer    *workpostgres.Authorizer
	hits          *workpostgres.CompatibilityRecorder
	query         *workquery.Service
	projections   *projectionapp.Service
	taskResources LegacyTaskResourceResolver
	services      httpServices
}

func NewPlatformHTTPRuntime(deps PlatformHTTPRuntimeDependencies) (*PlatformHTTPRuntime, error) {
	if deps.Generations == nil || deps.Scopes == nil || deps.Principals == nil ||
		deps.ProjectionResources == nil || deps.TaskReferences == nil ||
		deps.TaskResources == nil {
		return nil, fmt.Errorf("Work HTTP runtime dependencies are required")
	}
	principals, err := newPrincipalResolver(deps.Principals)
	if err != nil {
		return nil, err
	}
	unit := workpostgres.NewBinder()
	authorizer, err := workpostgres.NewAuthorizer(principals)
	if err != nil {
		return nil, err
	}
	events, err := newEventPublisher(deps.Catalogs)
	if err != nil {
		return nil, err
	}
	mutator, err := NewMutationExecutor(MutationDependencies{UnitOfWork: unit,
		Authorizer: authorizer, Events: events, Clock: runtimeClock{}, IDs: runtimeIDs{}})
	if err != nil {
		return nil, err
	}
	services, err := buildHTTPServices(mutator, principals, deps.ProjectionResources,
		deps.TaskReferences)
	if err != nil {
		return nil, err
	}
	reader, err := workpostgres.NewReader(principals)
	if err != nil {
		return nil, err
	}
	hits, err := workpostgres.NewCompatibilityRecorder()
	if err != nil {
		return nil, err
	}
	query, err := newWorkQueryRuntime(deps.Principals, deps.CursorKeys)
	if err != nil {
		return nil, err
	}
	projectionService, err := projectionapp.New(projectionapp.Dependencies{
		Mutator: mutator, Resources: deps.ProjectionResources})
	if err != nil {
		return nil, err
	}
	return &PlatformHTTPRuntime{resolver: deps.Generations, scopes: deps.Scopes,
		principals: deps.Principals, unit: unit, reader: reader, authorizer: authorizer,
		hits: hits, query: query, projections: projectionService,
		taskResources: deps.TaskResources, services: services}, nil
}

var _ HTTPApplication = (*PlatformHTTPRuntime)(nil)
