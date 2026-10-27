package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) Copy(ctx context.Context, scope ports.Scope,
	request ApplyRequest,
) (Receipt, error) {
	if service == nil || scope == nil || !scope.TenantID().Valid() ||
		!scope.ActorID().Valid() || !request.Valid() {
		return Receipt{}, ports.ErrInvalid
	}
	if err := service.dependencies.Authorizer.Authorize(
		ctx, scope, "work.v1_adoption.apply", request.RunID); err != nil {
		return Receipt{}, err
	}
	records, err := service.dependencies.Source.LoadLegacy(ctx, scope)
	if err != nil {
		return Receipt{}, err
	}
	plan, err := service.buildPlan(ctx, scope, request, records)
	if err != nil {
		return Receipt{}, err
	}
	if err = service.registerFormReferences(ctx, scope, &plan); err != nil {
		return Receipt{}, err
	}
	receipt, err := service.dependencies.Target.PersistPlan(ctx, scope, plan)
	if err != nil {
		return Receipt{}, err
	}
	if len(plan.Issues) > 0 {
		return receipt, ErrBlocked
	}
	return receipt, nil
}
