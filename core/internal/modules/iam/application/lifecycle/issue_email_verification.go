package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) IssueEmailVerification(
	ctx context.Context, scope platformtx.TenantScope,
	command EmailVerificationStartCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.IdentifierID == uuid.Nil || command.Token.Validate() != nil ||
		command.Delivery.Validate(scope.TenantID(), PurposeEmailVerification,
			command.Token) != nil || command.Delivery.Envelope.ResourceID != command.IdentifierID {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.IssueEmailVerification(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	change := changeFrom(command.Evidence, "iam.email_verification.start",
		"one_time_token", result)
	change.DeliveryArtifactRef = command.Delivery.Reference
	return service.finishTenant(ctx, scope, result, change)
}
