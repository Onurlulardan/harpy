package profilebackfill

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RecoveryService struct {
	store   RecoveryStore
	effects Effects
}

func NewRecoveryService(
	store RecoveryStore,
	effects Effects,
) (*RecoveryService, error) {
	if interfaceNil(store) || interfaceNil(effects) {
		return nil, ErrUnavailable
	}
	return &RecoveryService{store: store, effects: effects}, nil
}

func (service *RecoveryService) Start(
	ctx context.Context,
	scope platformtx.TenantScope,
	command RecoveryCreateCommand,
) (RecoveryResult, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		scope.ActorID() != command.Evidence.ActorID {
		return RecoveryResult{}, ErrInvalid
	}
	prepared, err := PrepareRecovery(command, scope.TenantID())
	if err != nil {
		return RecoveryResult{}, err
	}
	result, err := service.store.StartRecovery(ctx, scope, prepared)
	if err != nil {
		return RecoveryResult{}, err
	}
	if !validRecoveryResult(result, command.Plan) {
		return RecoveryResult{}, ErrUnavailable
	}
	if result.Replayed || prepared.FirstJob == nil {
		return result, nil
	}
	if err := service.effects.Enqueue(ctx, scope, *prepared.FirstJob); err != nil {
		return RecoveryResult{}, ErrEffects
	}
	return result, nil
}

func validRecoveryResult(result RecoveryResult, plan RecoveryPlan) bool {
	if result.PlanID != plan.ID() || result.Revision < 1 ||
		result.AuthorizationClock < 1 || !result.State.Valid() {
		return false
	}
	if result.Replayed {
		return true
	}
	expectedState := RecoveryPlanned
	if plan.RemainingSourceCount() == 0 {
		expectedState = RecoveryActivated
	}
	return result.State == expectedState
}
