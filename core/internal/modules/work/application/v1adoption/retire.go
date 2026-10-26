package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) Retire(ctx context.Context, scope ports.Scope,
	request RetireRequest,
) (Receipt, error) {
	if service == nil || scope == nil || !scope.TenantID().Valid() ||
		!scope.ActorID().Valid() || !request.Valid() {
		return Receipt{}, ports.ErrInvalid
	}
	if err := service.dependencies.Authorizer.Authorize(
		ctx, scope, "work.v1_adoption.retire", request.RunID); err != nil {
		return Receipt{}, err
	}
	evidence, err := service.dependencies.Target.LoadEvidence(ctx, scope, request.RunID)
	if err != nil {
		return Receipt{}, err
	}
	if !evidence.Accepted || !evidence.WriterFenced || evidence.CompatibilityHits != 0 {
		return Receipt{}, ErrBlocked
	}
	return service.dependencies.Target.SealLegacy(ctx, scope, request, evidence)
}
