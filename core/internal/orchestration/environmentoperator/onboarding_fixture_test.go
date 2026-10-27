package environmentoperator

import (
	"context"
	"errors"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type onboardingFixture struct {
	plan          tenantapi.OnboardingPlan
	state         tenantapi.OnboardingState
	revision      int64
	activeAttempt tenantapi.UUID
	accepts       int
	claims        int
	records       int
}

func (fixture *onboardingFixture) Accept(
	_ context.Context,
	plan tenantapi.OnboardingPlan,
) (tenantapi.OnboardingOperation, error) {
	fixture.accepts++
	if fixture.revision == 0 {
		fixture.plan, fixture.state, fixture.revision =
			plan, tenantapi.OnboardingAccepted, 1
	} else if fixture.plan.StepPlanSHA256() != plan.StepPlanSHA256() {
		return tenantapi.OnboardingOperation{}, tenantapi.ErrOnboardingConflict
	}
	return tenantapi.NewOnboardingOperation(
		fixture.plan, fixture.state, fixture.revision,
		fixture.plan.CreatedAt().Add(time.Duration(fixture.revision)*time.Microsecond),
	)
}

func (fixture *onboardingFixture) Claim(
	_ context.Context,
	request tenantapi.ClaimAttemptRequest,
) (tenantapi.AttemptLease, error) {
	fixture.claims++
	if request.ExpectedRevision != fixture.revision {
		return tenantapi.AttemptLease{}, tenantapi.ErrOnboardingConflict
	}
	if fixture.activeAttempt != request.AttemptID {
		fixture.activeAttempt = request.AttemptID
		fixture.revision++
		if fixture.state == tenantapi.OnboardingAccepted {
			fixture.state = tenantapi.OnboardingMigrating
		}
	}
	return tenantapi.NewAttemptLease(request.OperationID, request.AttemptID,
		int64(fixture.claims), fixture.revision,
		fixture.plan.CreatedAt().Add(10*time.Minute))
}

func (fixture *onboardingFixture) Record(
	_ context.Context,
	receipt tenantapi.OnboardingReceipt,
	expected int64,
) (tenantapi.OnboardingTransition, error) {
	if expected != fixture.revision || !receipt.IsValid() {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingConflict
	}
	fixture.records++
	fixture.revision++
	fixture.state = migrationNextState(fixture.plan.DeploymentMode())
	return tenantapi.OnboardingTransition{OperationID: fixture.plan.OperationID(),
		State: fixture.state, Revision: fixture.revision,
		UpdatedAt: receipt.OccurredAt.Add(time.Microsecond)}, nil
}

func (fixture *onboardingFixture) ReceiveAuthorityAck(
	_ context.Context,
	ack tenantapi.AuthorityAcknowledgement,
	expected int64,
) (tenantapi.OnboardingTransition, error) {
	if expected != fixture.revision || !ack.IsValid() {
		return tenantapi.OnboardingTransition{}, errors.New("ack rejected")
	}
	fixture.revision++
	fixture.state = tenantapi.OnboardingAuthorityAcknowledged
	return tenantapi.OnboardingTransition{OperationID: fixture.plan.OperationID(),
		State: fixture.state, Revision: fixture.revision,
		UpdatedAt: ack.VerifiedAt.Add(time.Microsecond)}, nil
}
