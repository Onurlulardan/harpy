package restquery

import (
	"context"
	"fmt"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const Owner = "work"

type CursorKeys = queryruntime.CursorKeys

func NewCursorKeys(currentID string, current [32]byte,
	previousID string, previous [32]byte,
) (CursorKeys, error) {
	return queryruntime.NewCursorKeys(currentID, current, previousID, previous)
}

type Service struct {
	engine   *queryruntime.Engine
	catalogs map[string]ResolvedCatalog
}

func New(principals workprincipalapi.Directory, catalogs CatalogPort,
	owner OwnerPort, keys CursorKeys,
) (*Service, error) {
	if principals == nil || catalogs == nil || owner == nil {
		return nil, fmt.Errorf("Work REST Query dependencies are required")
	}
	provided, err := catalogs()
	if err != nil {
		return nil, fmt.Errorf("load Work REST Query catalogs: %w", err)
	}
	resolved := make(map[string]ResolvedCatalog, len(provided))
	for _, catalog := range provided {
		if catalog.Validate() != nil || resolved[catalog.resource].Validate() == nil {
			return nil, fmt.Errorf("invalid Work REST Query catalog set")
		}
		resolved[catalog.resource] = catalog
	}
	for _, resource := range workQueryResources() {
		if resolved[resource].Validate() != nil {
			return nil, fmt.Errorf("missing Work REST Query catalog: %s", resource)
		}
	}
	if len(resolved) != len(workQueryResources()) {
		return nil, fmt.Errorf("unexpected Work REST Query catalog")
	}
	policy := &workQueryPolicy{catalogs: resolved, principals: principals}
	engine, err := queryruntime.NewExact(Owner, policy, owner, keys)
	if err != nil {
		return nil, fmt.Errorf("build Work REST Query runtime: %w", err)
	}
	return &Service{engine: engine, catalogs: resolved}, nil
}

func (service *Service) Query(ctx context.Context, scope platformtx.TenantScope,
	evidence Evidence, request restapi.QueryRequest,
) (queryruntime.Result, error) {
	if service == nil || service.engine == nil || !evidence.valid() || request.Validate() != nil {
		return queryruntime.Result{}, ErrInvalid
	}
	catalog := service.catalogs[evidence.resource]
	if catalog.Validate() != nil {
		return queryruntime.Result{}, ErrInvalid
	}
	return service.engine.ExecuteCatalog(withEvidence(ctx, evidence), scope, catalog.set, request)
}
