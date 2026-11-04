package environmentoperator

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type provisionFixture struct {
	onboarding         *onboardingFixture
	calls              int
	tenantID           uuid.UUID
	err                error
	preserveOnboarding bool
}

func (fixture *provisionFixture) Provision(_ context.Context, tenantID uuid.UUID) error {
	fixture.calls++
	fixture.tenantID = tenantID
	if fixture.err != nil {
		return fixture.err
	}
	if fixture.preserveOnboarding || fixture.onboarding == nil {
		return nil
	}
	if fixture.onboarding.state == tenantapi.OnboardingMigrating {
		fixture.onboarding.state = migrationNextState(fixture.onboarding.plan.DeploymentMode())
		fixture.onboarding.revision++
	}
	return nil
}
