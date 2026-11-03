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
		return 0, Resumable("tenant_migration.enqueue", "migration_enqueue_failed", "", err)
	}
	if err := validateMigrationJob(job, plan); err != nil {
		return 0, err
	}
	deadline := job.CreatedAt.Add(maximumMigrationWait)
	for {
		job, found, getErr := coordinator.migration.Get(ctx, uuid.UUID(plan.OperationID()))
		if getErr != nil {
			return 0, Resumable("tenant_migration.get", "migration_get_failed", "", getErr)
		}
		if !found || validateMigrationJob(job, plan) != nil {
			return 0, ErrFenced
		}
		switch job.State {
		case "succeeded":
			return coordinator.refreshMigration(ctx, plan, claim, job)
		case "failed":
			return 0, ErrFailed
		case "fenced", "reconcile_required", "cleanup_required":
			return 0, ErrFenced
		}
		if !job.DatabaseNow.Before(deadline) {
			return 0, Resumable("tenant_migration.wait_deadline", "migration_deadline",
				migrationDiagnostic(plan, claim, job, "", 0), nil)
		}
		wait := coordinator.poll
		if remaining := deadline.Sub(job.DatabaseNow); remaining < wait {
			wait = remaining
		}
		if err := coordinator.sleep(ctx, wait); err != nil {
			return 0, Resumable("tenant_migration.wait", "migration_wait_interrupted", "", err)
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
	job tenantapi.TenantMigrationJob,
) (int64, error) {
	operation, err := coordinator.onboarding.Accept(ctx, plan)
	if err != nil {
		return 0, Resumable("tenant_migration.refresh_onboarding", "onboarding_accept_failed",
			migrationDiagnostic(plan, claim, job, "", 0), err)
	}
	if operation.State() != migrationNextState(plan.DeploymentMode()) ||
		operation.Revision() <= claim.revision {
		return 0, Resumable("tenant_migration.refresh_onboarding", "onboarding_state_mismatch",
			migrationDiagnostic(plan, claim, job, operation.State(), operation.Revision()), nil)
	}
	return operation.Revision(), nil
}
