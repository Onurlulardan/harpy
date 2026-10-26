package profilebackfill

import (
	"context"
	"reflect"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Create(context.Context, platformtx.TenantScope, PreparedPlan) (Result, error)
}

type Effects interface {
	Enqueue(context.Context, platformtx.TenantScope, FirstJob) error
}

type Service struct {
	store   Store
	effects Effects
}

func NewService(store Store, effects Effects) (*Service, error) {
	if interfaceNil(store) || interfaceNil(effects) {
		return nil, ErrUnavailable
	}
	return &Service{store: store, effects: effects}, nil
}

func (service *Service) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	command CreateCommand,
) (Result, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return Result{}, ErrInvalid
	}
	prepared, err := prepare(command, scope.TenantID())
	if err != nil {
		return Result{}, err
	}
	result, err := service.store.Create(ctx, scope, prepared)
	if err != nil {
		return Result{}, err
	}
	if result.PlanID != command.Plan.ID || result.Revision < 1 ||
		result.AuthorizationClock < 1 {
		return Result{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if err := service.effects.Enqueue(ctx, scope, prepared.FirstJob); err != nil {
		return Result{}, ErrEffects
	}
	return result, nil
}

func interfaceNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
