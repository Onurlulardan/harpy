package environmentoperator

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) migrate(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
) (int64, error) {
	job, err := coordinator.migration.Enqueue(ctx, uuid.UUID(plan.OperationID()))
	if err != nil {
		return 0, ErrResumable
	}
	if err := validateMigrationJob(job, plan); err != nil {
		return 0, err
	}
	deadline := job.CreatedAt.Add(maximumMigrationWait)
	for {
		job, found, getErr := coordinator.migration.Get(ctx, uuid.UUID(plan.OperationID()))
		if getErr != nil {
			return 0, ErrResumable
		}
		if !found || validateMigrationJob(job, plan) != nil {
			return 0, ErrFenced
		}
		switch job.State {
		case "succeeded":
			return coordinator.refreshMigration(ctx, plan, claim)
		case "failed":
			return 0, ErrFailed
		case "fenced", "reconcile_required", "cleanup_required":
			return 0, ErrFenced
		}
		if !job.DatabaseNow.Before(deadline) {
			return 0, ErrResumable
		}
		wait := coordinator.poll
		if remaining := deadline.Sub(job.DatabaseNow); remaining < wait {
			wait = remaining
		}
		if err := coordinator.sleep(ctx, wait); err != nil {
			return 0, ErrResumable
		}
	}
}

func validateMigrationJob(job tenantapi.TenantMigrationJob,
	plan tenantapi.OnboardingPlan,
) error {
	if !job.IsValid() || job.OperationID != tenantapi.UUIDString(plan.OperationID()) ||
		job.TenantID != tenantapi.UUIDString(plan.TenantID()) ||
		job.SchemaName != plan.SchemaName() ||
		job.StepPlanSHA256 != plan.StepPlanSHA256() || job.CreatedAt.IsZero() {
		return ErrFenced
	}
	return nil
}

func (coordinator *Coordinator) refreshMigration(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
) (int64, error) {
	operation, err := coordinator.onboarding.Accept(ctx, plan)
	if err != nil {
		return 0, ErrResumable
	}
	if operation.State() != migrationNextState(plan.DeploymentMode()) ||
		operation.Revision() <= claim.revision {
		return 0, ErrResumable
	}
	return operation.Revision(), nil
}
