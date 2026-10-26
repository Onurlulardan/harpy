package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) Reopen(ctx context.Context, scope ports.Scope,
	request ReverseRequest,
) (Receipt, error) {
	if service == nil || scope == nil || !scope.TenantID().Valid() ||
		!scope.ActorID().Valid() || !request.Valid() || request.ApprovedBy == scope.ActorID() {
		return Receipt{}, ports.ErrInvalid
	}
	if err := service.dependencies.Authorizer.Authorize(
		ctx, scope, "work.v1_adoption.reverse", request.RunID); err != nil {
		return Receipt{}, err
	}
	evidence, err := service.dependencies.Target.LoadEvidence(ctx, scope, request.RunID)
	if err != nil {
		return Receipt{}, err
	}
	if !evidence.Accepted || !evidence.WriterFenced || evidence.SourceCount < 0 ||
		evidence.TargetCount < 0 || evidence.SourceSHA256 == [32]byte{} ||
		evidence.TargetSHA256 == [32]byte{} {
		return Receipt{}, ErrBlocked
	}
	return service.dependencies.Target.RevertPlan(ctx, scope, request, evidence)
}
