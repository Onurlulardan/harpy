package customerprovisioning

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) beginSchema(
	ctx context.Context, command Command, plan tenantapi.OnboardingPlan,
) error {
	schema := coordinator.dependencies.Schema
	if schema == nil {
		return nil
	}
	err := schema.Begin(
		ctx, uuid.UUID(plan.TenantID()), plan.SchemaName(),
		command.Customer.Name, command.Customer.Slug,
	)
	if err != nil {
		return ErrUnavailable
	}
	return nil
}

func (coordinator *Coordinator) activateSchema(
	ctx context.Context, plan tenantapi.OnboardingPlan,
) error {
	schema := coordinator.dependencies.Schema
	if schema == nil {
		return nil
	}
	if err := schema.Activate(ctx, uuid.UUID(plan.TenantID())); err != nil {
		return ErrUnavailable
	}
	return nil
}
