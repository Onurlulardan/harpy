package protectedadmin

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"reflect"
)

type Commands struct {
	store rbacports.ProtectedAdminStore[transaction.TenantScope]
	sink  rbacports.ChangeSink[transaction.TenantScope]
}

func NewCommands(store rbacports.ProtectedAdminStore[transaction.TenantScope],
	sink rbacports.ChangeSink[transaction.TenantScope]) (*Commands, error) {
	if nilPort(store) || nilPort(sink) {
		return nil, rbacdomain.ErrStorage
	}
	return &Commands{store: store, sink: sink}, nil
}

func mutationScope(ctx context.Context, scope transaction.TenantScope,
	evidence rbacdomain.MutationEvidence) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil &&
		scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func queryScope(ctx context.Context, scope transaction.TenantScope,
	evidence rbacdomain.QueryEvidence) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.SnapshotQuery) == nil &&
		scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Pointer && v.IsNil()
}
