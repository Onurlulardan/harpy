package worker

import (
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func TestCustomerActivationRequiresOwnerlessGenerationOneAuthorityAck(t *testing.T) {
	plan := customerActivationPlan(t)
	claim := tenancy.CustomerWorkerClaim{Plan: plan,
		Customer: tenancy.CustomerWorkerSnapshot{Name: "Acme", Slug: "acme",
			FirstAdmin: tenancy.CustomerWorkerFirstAdminSnapshot{DisplayName: "Ada", Email: "ada@acme.test"}},
		State: tenancy.OnboardingAuthorityAcknowledged, Revision: 4,
		LeaseTokenSHA256: strings.Repeat("e", 64), LeaseUntil: time.Now().Add(time.Minute),
	}
	authority := plan.Authority()
	binding := tenancy.OperatorActivationBinding{
		OperationID: plan.OperationID(), EnvironmentID: plan.EnvironmentID(),
		TenantID: plan.TenantID(), FirstUserID: plan.FirstUserID(),
		Kind: "saas_customer_tenant_provision", DeploymentMode: "saas",
		BindingPurpose: "saas_customer", AuthorityScopeKind: "tenant",
		AuthorityScopeID: plan.TenantID(), AuthorityGeneration: 1,
		AuthorityRegistryBindingID: authority.RegistryBindingID,
		AuthorityStateSHA256:       authority.StateSHA256,
		AuthorityProposalSHA256:    authority.ProposalSHA256,
		AuthorityAckSHA256:         strings.Repeat("d", 64),
	}
	resources := tenancy.OperatorActivationResources{
		BindOwner: func(tenancy.OperatorActivationOwner) (
			tenancy.OperatorActivationExecutor, error,
		) {
			return tenancy.OperatorActivationExecutor{}, nil
		},
		Consume: func(tenancy.OperatorActivationCommand) error { return nil },
	}
	if !validCustomerActivation(claim, binding, resources) {
		t.Fatal("valid ownerless generation-one customer authority ack was rejected")
	}
	binding.InstanceOwnerBindingID = [16]byte{1}
	if validCustomerActivation(claim, binding, resources) {
		t.Fatal("customer activation accepted an instance owner binding")
	}
	binding.InstanceOwnerBindingID = tenancy.NilUUID
	binding.AuthorityGeneration = 2
	if validCustomerActivation(claim, binding, resources) {
		t.Fatal("customer activation accepted authority generation two")
	}
	binding.AuthorityGeneration = 1
	binding.AuthorityAckSHA256 = ""
	if validCustomerActivation(claim, binding, resources) {
		t.Fatal("customer activation accepted a missing protected authority ack")
	}
}

func customerActivationPlan(t *testing.T) tenancy.OnboardingPlan {
	t.Helper()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	id := func(value string) tenancy.UUID {
		parsed, err := tenancy.ParseUUID(value)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	tenantID := id("71000000-0000-7000-8000-000000000001")
	schemaName, err := tenancy.OnboardingSchemaName(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := tenancy.NewOnboardingPlan(tenancy.OnboardingPlanInput{
		OperationID:   id("72000000-0000-7000-8000-000000000001"),
		EnvironmentID: id("73000000-0000-7000-8000-000000000001"), TenantID: tenantID,
		FirstUserID:             id("74000000-0000-7000-8000-000000000001"),
		RootOrganizationID:      id("75000000-0000-7000-8000-000000000001"),
		ProtectedAdminBindingID: id("76000000-0000-7000-8000-000000000001"),
		Kind:                    tenancy.OnboardingSaaSCustomerProvision,
		DeploymentMode:          tenancy.DeploymentModeSaaS,
		BindingPurpose:          tenancy.TenantBindingSaaSCustomer,
		SchemaName:              schemaName, IdempotencyKey: "customer-acme",
		RequestSHA256: strings.Repeat("1", 64), ManifestSHA256: strings.Repeat("2", 64),
		Offering: tenancy.BaseOfferingPin{VersionID: id("77000000-0000-7000-8000-000000000001"),
			Revision: 1, SHA256: strings.Repeat("3", 64), EffectiveAt: now},
		Authority: tenancy.ProtectedAuthorityPlan{
			RegistryBindingID: id("78000000-0000-7000-8000-000000000001"),
			ScopeKind:         tenancy.AuthorityScopeTenant, ScopeID: tenantID, Generation: 1,
			StateSHA256: strings.Repeat("4", 64), ProposalSHA256: strings.Repeat("5", 64)},
		CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
