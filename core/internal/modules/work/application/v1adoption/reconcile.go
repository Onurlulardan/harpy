package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) Reconcile(ctx context.Context, scope ports.Scope,
	request ReconcileRequest,
) (Evidence, error) {
	if service == nil || scope == nil || !scope.TenantID().Valid() ||
		!scope.ActorID().Valid() || !request.Valid() {
		return Evidence{}, ports.ErrInvalid
	}
	if err := service.dependencies.Authorizer.Authorize(
		ctx, scope, "work.v1_adoption.reconcile", request.RunID); err != nil {
		return Evidence{}, err
	}
	records, err := service.dependencies.Source.LoadLegacy(ctx, scope)
	if err != nil {
		return Evidence{}, err
	}
	plan, err := service.buildPlan(ctx, scope, request, records)
	if err != nil {
		return Evidence{}, err
	}
	if err = service.registerFormReferences(ctx, scope, &plan); err != nil {
		return Evidence{}, err
	}
	evidence, err := service.dependencies.Target.ComparePlan(ctx, scope, plan)
	if err != nil {
		return Evidence{}, err
	}
	if len(plan.Issues) > 0 || !evidence.Accepted ||
		evidence.SourceCount != plan.SourceCount || evidence.SourceSHA256 != plan.SourceSHA256 ||
		evidence.TargetCount != int64(len(plan.Records)) ||
		evidence.TargetSHA256 != targetDigest(plan.Records) {
		return evidence, ErrBlocked
	}
	return evidence, nil
}
