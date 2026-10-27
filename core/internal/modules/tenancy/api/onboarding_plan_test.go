package api

import (
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestNewOnboardingPlanPinsStableEnvironmentIdentity(t *testing.T) {
	input := validPlanInput(t, DeploymentModeOnPremise)
	plan, err := NewOnboardingPlan(input)
	if err != nil || !plan.IsValid() || !validOnboardingDigest(plan.StepPlanSHA256()) {
		t.Fatalf("plan = %#v, %v", plan, err)
	}
	duplicate, err := NewOnboardingPlan(input)
	if err != nil || duplicate.StepPlanSHA256() != plan.StepPlanSHA256() {
		t.Fatalf("deterministic plan digest = %q, %v", duplicate.StepPlanSHA256(), err)
	}
	input.FirstUserID = randomOnboardingUUID(t)
	changed, _ := NewOnboardingPlan(input)
	if changed.StepPlanSHA256() == plan.StepPlanSHA256() {
		t.Fatal("stable owner identity was not pinned into the plan")
	}
}

func TestNewOnboardingPlanRejectsModeAndSigningDrift(t *testing.T) {
	input := validPlanInput(t, DeploymentModeOnPremise)
	input.RecoverySigning = RecoverySigningPin{}
	if _, err := NewOnboardingPlan(input); err == nil {
		t.Fatal("on-premise plan accepted without recovery signing output")
	}
	input = validPlanInput(t, DeploymentModeSaaS)
	input.Kind = OnboardingSaaSCustomerProvision
	input.BindingPurpose = TenantBindingSaaSCustomer
	input.InstanceOwnerBindingID = NilUUID
	input.Authority.ScopeKind = AuthorityScopeTenant
	input.Authority.ScopeID = input.TenantID
	if _, err := NewOnboardingPlan(input); err != nil {
		t.Fatalf("valid SaaS customer plan rejected: %v", err)
	}
	input.Authority.ScopeID = randomOnboardingUUID(t)
	if _, err := NewOnboardingPlan(input); err == nil {
		t.Fatal("wrong protected authority scope accepted")
	}
}

func validPlanInput(t *testing.T, mode DeploymentMode) OnboardingPlanInput {
	t.Helper()
	tenantID := randomOnboardingUUID(t)
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	input := OnboardingPlanInput{OperationID: randomOnboardingUUID(t),
		EnvironmentID: randomOnboardingUUID(t), TenantID: tenantID,
		FirstUserID: randomOnboardingUUID(t), RootOrganizationID: randomOnboardingUUID(t),
		ProtectedAdminBindingID: randomOnboardingUUID(t),
		InstanceOwnerBindingID:  randomOnboardingUUID(t),
		Kind:                    OnboardingEnvironmentInitialize, DeploymentMode: mode,
		BindingPurpose: TenantBindingSaaSControl, SchemaName: onboardingSchemaName(tenantID),
		IdempotencyKey: "initialize", RequestSHA256: strings.Repeat("1", 64),
		ManifestSHA256: strings.Repeat("2", 64), CreatedAt: now,
		Offering: BaseOfferingPin{VersionID: randomOnboardingUUID(t), Revision: 1,
			SHA256: strings.Repeat("3", 64), EffectiveAt: now},
		Authority: ProtectedAuthorityPlan{RegistryBindingID: randomOnboardingUUID(t),
			ScopeKind: AuthorityScopeEnvironment, Generation: 1,
			StateSHA256: strings.Repeat("6", 64), ProposalSHA256: strings.Repeat("7", 64)}}
	input.Authority.ScopeID = input.EnvironmentID
	if mode == DeploymentModeOnPremise {
		input.BindingPurpose = TenantBindingOnPremisePrimary
		input.Offering.Source = OfferingSourcePin{Sequence: 1,
			SHA256: strings.Repeat("4", 64), Present: true}
		input.RecoverySigning = RecoverySigningPin{Generation: 1,
			TrustBundleSHA256: strings.Repeat("5", 64), Present: true}
		input.Authority.Generation = 0
		input.Authority.StateSHA256 = ""
		input.Authority.ProposalSHA256 = ""
	}
	return input
}

func randomOnboardingUUID(t *testing.T) UUID {
	t.Helper()
	var result UUID
	if _, err := rand.Read(result[:]); err != nil {
		t.Fatal(err)
	}
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}
