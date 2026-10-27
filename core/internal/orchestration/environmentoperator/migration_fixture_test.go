package environmentoperator

import (
	"context"
	"strings"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type migrationFixture struct {
	onboarding   *onboardingFixture
	states       []string
	index        int
	ready        bool
	readyAfter   int
	readyCalls   int
	enqueues     int
	gets         int
	createdAt    time.Time
	databaseStep time.Duration
}

func (fixture *migrationFixture) CurrentReady(context.Context) (
	tenantapi.TenantMigrationReady, bool, error,
) {
	fixture.readyCalls++
	if !fixture.ready &&
		(fixture.readyAfter < 1 || fixture.readyCalls < fixture.readyAfter) {
		return tenantapi.TenantMigrationReady{}, false, nil
	}
	plan := fixture.onboarding.plan
	return tenantapi.TenantMigrationReady{
		ReleaseID:      "20000000-0000-4000-8000-000000000001",
		EnvironmentID:  tenantapi.UUIDString(plan.EnvironmentID()),
		DeploymentMode: string(plan.DeploymentMode()), RuntimeGeneration: 13,
		DatabaseGeneration: 15, MigrationManifestSHA256: strings.Repeat("9", 64),
		ReadyUntil: fixture.createdAt.Add(10 * time.Minute),
	}, true, nil
}

func (fixture *migrationFixture) Enqueue(_ context.Context,
	operationID uuid.UUID,
) (tenantapi.TenantMigrationJob, error) {
	fixture.enqueues++
	return fixture.job(operationID, "pending", fixture.createdAt), nil
}

func (fixture *migrationFixture) Get(_ context.Context,
	operationID uuid.UUID,
) (tenantapi.TenantMigrationJob, bool, error) {
	fixture.gets++
	state := "pending"
	if len(fixture.states) > 0 {
		state = fixture.states[fixture.index]
		if fixture.index < len(fixture.states)-1 {
			fixture.index++
		}
	}
	if state == "succeeded" &&
		fixture.onboarding.state == tenantapi.OnboardingMigrating {
		fixture.onboarding.state = migrationNextState(
			fixture.onboarding.plan.DeploymentMode(),
		)
		fixture.onboarding.revision++
	}
	databaseNow := fixture.createdAt.Add(time.Duration(fixture.gets) * fixture.databaseStep)
	return fixture.job(operationID, state, databaseNow), true, nil
}

func (fixture *migrationFixture) job(operationID uuid.UUID,
	state string,
	databaseNow time.Time,
) tenantapi.TenantMigrationJob {
	plan := fixture.onboarding.plan
	job := tenantapi.TenantMigrationJob{
		JobID:       "30000000-0000-4000-8000-000000000001",
		OperationID: operationID.String(), State: state,
		TenantID: tenantapi.UUIDString(plan.TenantID()), SchemaName: plan.SchemaName(),
		StepPlanSHA256:    plan.StepPlanSHA256(),
		ReleaseID:         "20000000-0000-4000-8000-000000000001",
		RuntimeGeneration: 13, DatabaseGeneration: 15,
		MigrationManifestSHA256: strings.Repeat("9", 64),
		CreatedAt:               fixture.createdAt, DatabaseNow: databaseNow,
	}
	if state == "succeeded" {
		job.CompletionReceiptID = "40000000-0000-4000-8000-000000000001"
		job.CompletionReceiptSHA256 = strings.Repeat("a", 64)
		job.CompletionEvidenceSHA256 = strings.Repeat("b", 64)
		job.CompletedAt = databaseNow
	}
	return job
}
