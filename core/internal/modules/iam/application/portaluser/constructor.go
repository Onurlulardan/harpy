package portaluser

import (
	"context"
	"reflect"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Service struct{ store Store }

func NewService(store Store) (*Service, error) {
	if nilStore(store) {
		return nil, ErrUnavailable
	}
	return &Service{store: store}, nil
}

func nilStore(value Store) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

func scopeUsable(
	ctx context.Context, scope platformtx.TenantScope, evidence Evidence, mode platformtx.ScopeMode,
) bool {
	return ctx != nil && ctx.Err() == nil && evidence.Valid() &&
		!platformtx.NetworkAllowed(ctx) && platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, mode) == nil && scope.ActorID() == evidence.ActorID
}

func validResult(result MutationResult, id interface{ String() string }) bool {
	return result.Valid() && result.ID.String() == id.String()
}
