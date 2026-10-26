package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func EnvironmentActivationFinalizer(
	input EnvironmentFinalizerInput,
) (tenancy.OperatorActivationCallback, error) {
	if !validEnvironmentFinalizerInput(input) {
		return nil, ErrEnvironmentFinalizerInvalid
	}
	return func(
		ctx context.Context,
		binding tenancy.OperatorActivationBinding,
		resources tenancy.OperatorActivationResources,
	) error {
		if ctx == nil || ctx.Err() != nil || binding.Kind != "environment_initialize" ||
			resources.BindOwner == nil || resources.Consume == nil {
			return ErrEnvironmentFinalizerInvalid
		}
		if err := finalizeTenantAndIdentity(ctx, binding, resources, input); err != nil {
			return fmt.Errorf("finalize tenant identity: %w", err)
		}
		if err := finalizeBillingAndLayout(ctx, binding, resources, input); err != nil {
			return fmt.Errorf("finalize tenant baseline: %w", err)
		}
		if err := finalizePlatformEvidence(ctx, binding, resources); err != nil {
			return fmt.Errorf("finalize platform evidence: %w", err)
		}
		return nil
	}, nil
}

func consume(
	resources tenancy.OperatorActivationResources,
	command tenancy.OperatorActivationCommand,
) error {
	if err := resources.Consume(command); err != nil {
		return fmt.Errorf("consume %s: %w", command, err)
	}
	return nil
}
