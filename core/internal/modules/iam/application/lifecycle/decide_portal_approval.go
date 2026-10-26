package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) DecidePortalApproval(
	ctx context.Context, scope platformtx.TenantScope, command PortalDecisionCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.BindingID == uuid.Nil || command.ExpectedRevision < 1 ||
		!command.Decision.Valid() {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.DecidePortalApproval(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.portal_approval.decide", "portal_binding", result))
}
