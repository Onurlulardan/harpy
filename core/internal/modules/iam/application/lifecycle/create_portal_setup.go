package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) CreatePortalSetup(
	ctx context.Context, scope platformtx.TenantScope,
	command PortalSetupStartCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.BindingID == uuid.Nil || command.Token.Validate() != nil ||
		command.Delivery.Validate(scope.TenantID(), PurposePortalSetup,
			command.Token) != nil || command.Delivery.Envelope.ResourceID != command.BindingID {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.CreatePortalSetup(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	change := changeFrom(command.Evidence, "iam.portal_setup.start",
		"one_time_token", result)
	change.DeliveryArtifactRef = command.Delivery.Reference
	return service.finishTenant(ctx, scope, result, change)
}
