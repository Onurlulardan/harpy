package query

import (
	"context"
	"encoding/json"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ExecuteRequest struct {
	QueryID, VersionID uuid.UUID
	Parameters         []byte
}

func (application *Application) Execute(
	ctx context.Context, scope platformtx.TenantScope, request ExecuteRequest,
) (ExecutionResult, error) {
	parameters, err := NewDocument(request.Parameters)
	if !validUUIDv7(request.QueryID) || !validUUIDv7(request.VersionID) || err != nil {
		return ExecutionResult{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeSnapshot,
		OperationExecute, PermissionExecute)
	if err != nil {
		return ExecutionResult{}, err
	}
	return application.executeBound(ctx, scope, store, identity, request, parameters)
}

func (application *Application) executeBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request ExecuteRequest, parameters Document,
) (ExecutionResult, error) {
	snapshot, err := store.LoadVersion(ctx, request.QueryID, request.VersionID)
	if err != nil || snapshot.VersionID != request.VersionID ||
		snapshot.VersionState != analyticsquery.VersionPublished || validateSnapshot(snapshot) != nil {
		return ExecutionResult{}, ErrConflict
	}
	dependencies, err := application.resolveExecutionDependencies(
		ctx, scope, store, request.QueryID, request.VersionID,
	)
	if err != nil {
		return ExecutionResult{}, err
	}
	binding, err := application.contracts.BindExecution(
		ctx, scope, snapshot, dependencies, parameters,
	)
	if err != nil || !validExecutionBinding(binding) {
		return ExecutionResult{}, ErrUnavailable
	}
	cacheRequest := CacheRequest{TenantID: identity.TenantID,
		QueryID: request.QueryID, QueryVersionID: request.VersionID, Binding: binding,
		QuerySHA256:      digestDocument(snapshot.QueryDocument),
		ParametersSHA256: parameters.SHA256()}
	if snapshot.CachePolicy == analyticspublication.CacheRevisionBound {
		if body, hit, cacheErr := application.cache.Read(ctx, cacheRequest); cacheErr == nil && hit && validResultBody(body) {
			return ExecutionResult{Body: body}, nil
		}
	}
	result, err := application.contracts.Run(
		ctx, scope, snapshot, dependencies, parameters,
	)
	if err != nil || !validResultBody(result.Body) {
		return ExecutionResult{}, ErrUnavailable
	}
	if snapshot.CachePolicy == analyticspublication.CacheRevisionBound {
		now := application.clock.Now()
		if validTime(now) {
			_ = application.cache.Write(ctx, cacheRequest, result.Body, now.Add(300*time.Second))
		}
	}
	return result, nil
}

func validExecutionBinding(value ExecutionBinding) bool {
	return value.ReleaseGeneration > 0 && value.AuthorizationSHA256 != [32]byte{} &&
		value.PredicateSHA256 != [32]byte{} && value.RevisionsSHA256 != [32]byte{}
}

func digestDocument(raw []byte) [32]byte {
	document, err := NewDocument(raw)
	if err != nil {
		return [32]byte{}
	}
	return document.SHA256()
}

func validResultBody(value []byte) bool {
	return len(value) > 0 && len(value) <= 1<<20 && json.Valid(value)
}
