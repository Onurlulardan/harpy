package customerprovisioning

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

var provisionNow = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return provisionNow }

type policyFixture struct {
	policy tenantapi.DeploymentPolicy
	target tenantapi.PrimaryTenantTarget
}

func (fixture policyFixture) Snapshot() (tenantapi.DeploymentPolicy, error) {
	return fixture.policy, nil
}
func (fixture policyFixture) PrimaryTarget() (tenantapi.PrimaryTenantTarget, error) {
	return fixture.target, nil
}

type accessFixture struct {
	calls        int
	tamperStepUp bool
}

func (fixture *accessFixture) Verify(_ context.Context, request AccessRequest) (AccessDecision, error) {
	fixture.calls++
	decision := AccessDecision{SubjectID: uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		ControlTenantID: request.ControlTenantID, DeploymentMode: tenantapi.DeploymentModeSaaS,
		Capability: request.Capability}
	if request.Capability == CapabilityCreate {
		decision.StepUpAt = provisionNow.Add(-time.Minute)
		decision.StepUpPurpose = ProvisionStepUpPurpose(request)
	}
	if request.Capability == CapabilityAdminSetupReissue {
		decision.StepUpAt = provisionNow.Add(-time.Minute)
		decision.StepUpPurpose = AdminSetupReissueStepUpPurpose(request)
		if fixture.tamperStepUp {
			decision.StepUpPurpose[0] ^= 0xff
		}
	}
	return decision, nil
}

type offeringFixture struct{}

func (offeringFixture) ResolveSaaSFreeBase(context.Context) (tenantapi.BaseOfferingPin, error) {
	return tenantapi.BaseOfferingPin{VersionID: tenantapi.UUID(uuid.MustParse(
		"019b0000-0000-7000-8000-000000000201")), Revision: 1,
		SHA256: strings.Repeat("a", 64), EffectiveAt: provisionNow}, nil
}

type authorityFixture struct{}

func (authorityFixture) PrepareInitialTenant(_ context.Context, request AuthorityRequest) (
	tenantapi.ProtectedAuthorityPlan, error,
) {
	return tenantapi.ProtectedAuthorityPlan{
		RegistryBindingID: tenantapi.UUID(uuid.MustParse("40000000-0000-4000-8000-000000000001")),
		ScopeKind:         tenantapi.AuthorityScopeTenant, ScopeID: tenantapi.UUID(request.TenantID),
		Generation: 1, StateSHA256: strings.Repeat("b", 64),
		ProposalSHA256: strings.Repeat("c", 64),
	}, nil
}

type setupFixture struct{ calls int }

func (fixture *setupFixture) PrepareInitialAdmin(
	_ context.Context, request SetupRequest,
) error {
	fixture.calls++
	if request.OperationID == uuid.Nil || request.TenantID == uuid.Nil ||
		request.UserID == uuid.Nil || request.Email == "" {
		return ErrInvalid
	}
	return nil
}

func (fixture *setupFixture) ReissueInitialAdmin(
	_ context.Context, request SetupReissueRequest,
) error {
	fixture.calls++
	if request.OperationID == uuid.Nil || request.ReissueOperationID == uuid.Nil ||
		request.TenantID == uuid.Nil || request.UserID == uuid.Nil ||
		request.SchemaName == "" || request.Email == "" {
		return ErrInvalid
	}
	return nil
}

type storeFixture struct {
	plan               tenantapi.OnboardingPlan
	accepted           AcceptedRequest
	operation          Operation
	reissueOperationID uuid.UUID
}

func (store *storeFixture) Create(_ context.Context, plan tenantapi.OnboardingPlan,
	accepted AcceptedRequest) (Operation, error) {
	if store.plan.IsValid() && store.plan.RequestSHA256() != plan.RequestSHA256() {
		return Operation{}, ErrConflict
	}
	store.plan, store.accepted = plan, accepted
	store.operation = Operation{OperationID: uuid.UUID(plan.OperationID()),
		TenantID: uuid.UUID(plan.TenantID()), RequestedBySubjectID: accepted.RequestedBySubjectID,
		State: tenantapi.OnboardingMigrating, SetupDeliveryState: "pending",
		CreatedAt: provisionNow, UpdatedAt: provisionNow}
	return store.operation, nil
}

func (store *storeFixture) Read(_ context.Context, operationID, subjectID uuid.UUID) (Operation, error) {
	if store.operation.OperationID != operationID || store.operation.RequestedBySubjectID != subjectID {
		return Operation{}, ErrNotFound
	}
	return store.operation, nil
}

func (store *storeFixture) LoadSetupReissueTarget(
	_ context.Context, tenantID uuid.UUID,
) (SetupReissueTarget, error) {
	if store.operation.TenantID != tenantID {
		return SetupReissueTarget{}, ErrNotFound
	}
	return SetupReissueTarget{Operation: store.operation, SchemaName: "t_acme",
		ReissueOperationID: store.reissueOperationID,
		UserID:             uuid.MustParse("50000000-0000-4000-8000-000000000001"),
		Email:              "ada@acme.test"}, nil
}

func TestProvisionIsStableOwnerlessAndReadAuthorized(t *testing.T) {
	coordinator, access, store := provisionFixture(t, tenantapi.DeploymentModeSaaS)
	command := validCommand(CallerHTTP)
	first, err := coordinator.Create(context.Background(), command)
	if err != nil || first.State != tenantapi.OnboardingMigrating || access.calls != 1 {
		t.Fatalf("first = %#v, calls=%d, err=%v", first, access.calls, err)
	}
	if store.plan.InstanceOwnerBindingID() != tenantapi.NilUUID ||
		store.plan.Authority().Generation != 1 ||
		store.plan.BindingPurpose() != tenantapi.TenantBindingSaaSCustomer {
		t.Fatal("customer plan gained owner authority or lost tenant generation one")
	}
	second, err := coordinator.Create(context.Background(), command)
	if err != nil || second.OperationID != first.OperationID ||
		second.TenantID != first.TenantID {
		t.Fatalf("replay = %#v, %v", second, err)
	}
	read, err := coordinator.Read(context.Background(), ReadQuery{
		Caller: CallerHTTP, OperationID: first.OperationID,
	})
	if err != nil || read.OperationID != first.OperationID {
		t.Fatalf("read = %#v, %v", read, err)
	}
}

func TestProvisionModeGuardPrecedesEveryCapabilityBoundary(t *testing.T) {
	for _, caller := range []CallerKind{CallerHTTP, CallerInternal} {
		t.Run(string(caller), func(t *testing.T) {
			coordinator, access, _ := provisionFixture(t, tenantapi.DeploymentModeOnPremise)
			_, err := coordinator.Create(context.Background(), validCommand(caller))
			if !errors.Is(err, ErrModeDenied) || access.calls != 0 {
				t.Fatalf("err=%v capability calls=%d", err, access.calls)
			}
		})
	}
}

func TestProvisionConflictingPayloadReusesNoLocation(t *testing.T) {
	coordinator, _, _ := provisionFixture(t, tenantapi.DeploymentModeSaaS)
	command := validCommand(CallerHTTP)
	if _, err := coordinator.Create(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	command.Customer.Name = "Different"
	if _, err := coordinator.Create(context.Background(), command); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflict err = %v", err)
	}
}
