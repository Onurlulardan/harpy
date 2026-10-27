package resource

import (
	"context"
	"reflect"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Commands struct {
	store rbacports.ResourceStore[transaction.TenantScope]
	sink  rbacports.ChangeSink[transaction.TenantScope]
}

func NewCommands(
	store rbacports.ResourceStore[transaction.TenantScope],
	sink rbacports.ChangeSink[transaction.TenantScope],
) (*Commands, error) {
	if nilPort(store) || nilPort(sink) {
		return nil, rbacdomain.ErrStorage
	}
	return &Commands{store: store, sink: sink}, nil
}

func (commands *Commands) finish(
	ctx context.Context, scope transaction.TenantScope,
	evidence rbacdomain.MutationEvidence, result rbacdomain.MutationResult,
	commandKey string, resultCode string,
) (rbacdomain.MutationResult, error) {
	if result.Validate(result.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	if result.Replayed {
		return result, nil
	}
	change := rbacdomain.NewChange(result, evidence, commandKey, resultCode)
	if change.Validate() != nil || commands.sink.WriteChange(ctx, scope, change) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrSink
	}
	return result, nil
}

func mutationScope(
	ctx context.Context, scope transaction.TenantScope,
	evidence rbacdomain.MutationEvidence,
) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil &&
		scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func queryScope(
	ctx context.Context, scope transaction.TenantScope,
	evidence rbacdomain.QueryEvidence,
) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.SnapshotQuery) == nil &&
		scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
