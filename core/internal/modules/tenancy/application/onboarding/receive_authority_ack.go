package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) ReceiveAuthorityAck(
	ctx context.Context,
	ack tenantapi.AuthorityAcknowledgement,
	expectedRevision int64,
) (tenantapi.OnboardingTransition, error) {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil ||
		!ack.IsValid() || expectedRevision < 1 {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingInvalid
	}
	transition, err := coordinator.store.ReceiveAuthorityAck(ctx, ack, expectedRevision)
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("receive authority acknowledgement: %w", err)
	}
	if !transition.IsValid() || transition.OperationID != ack.OperationID {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrAuthorityCASConflict
	}
	return transition, nil
}
