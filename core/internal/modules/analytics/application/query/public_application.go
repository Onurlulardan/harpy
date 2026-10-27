package query

import (
	"context"
	"fmt"
	"reflect"

	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PublicRootResolver interface {
	QueryRoot(context.Context, platformtx.TenantScope,
		uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

type PublicApplication struct {
	binder       StoreBinder
	dependencies DependencyResolver
	contracts    ContractPort
	roots        PublicRootResolver
}

func NewPublicApplication(
	binder StoreBinder, dependencies DependencyResolver,
	contracts ContractPort, roots PublicRootResolver,
) (*PublicApplication, error) {
	for _, value := range []any{binder, dependencies, contracts, roots} {
		if publicNil(value) {
			return nil, fmt.Errorf("analytics public query dependencies are invalid")
		}
	}
	return &PublicApplication{binder: binder, dependencies: dependencies,
		contracts: contracts, roots: roots}, nil
}

func (application *PublicApplication) Run(
	ctx context.Context, scope platformtx.TenantScope,
	releaseID, versionID uuid.UUID, parameterBytes []byte,
) (ExecutionResult, error) {
	parameters, parameterErr := NewDocument(parameterBytes)
	if application == nil || parameterErr != nil || !validUUIDv7(releaseID) ||
		!validUUIDv7(versionID) || ctx == nil || ctx.Err() != nil ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		platformtx.RequireOwner(scope, "analytics") != nil ||
		platformtx.RequireOwner(scope, "forms") != nil ||
		scope.OperationID() != "analytics.publication.execute" {
		return ExecutionResult{}, ErrInvalid
	}
	queryID, err := application.roots.QueryRoot(
		ctx, scope, releaseID, versionID)
	if err != nil || !validUUIDv7(queryID) {
		return ExecutionResult{}, ErrConflict
	}
	store, _, err := application.binder.Bind(ctx, scope, ScopeSnapshot)
	if err != nil || publicNil(store) {
		return ExecutionResult{}, ErrUnavailable
	}
	snapshot, err := store.LoadVersion(ctx, queryID, versionID)
	if err != nil || snapshot.QueryID != queryID || snapshot.VersionID != versionID ||
		snapshot.VersionState != analyticsquery.VersionPublished || validateSnapshot(snapshot) != nil {
		return ExecutionResult{}, ErrConflict
	}
	dependencies, err := application.resolveDependencies(
		ctx, scope, store, queryID, versionID)
	if err != nil {
		return ExecutionResult{}, err
	}
	binding, err := application.contracts.BindExecution(
		ctx, scope, snapshot, dependencies, parameters)
	if err != nil || !validExecutionBinding(binding) {
		return ExecutionResult{}, ErrUnavailable
	}
	result, err := application.contracts.Run(
		ctx, scope, snapshot, dependencies, parameters)
	if err != nil || !validResultBody(result.Body) {
		return ExecutionResult{}, ErrUnavailable
	}
	return result, nil
}

func publicNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
