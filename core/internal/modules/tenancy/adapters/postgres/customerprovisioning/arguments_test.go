package customerprovisioning

import (
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
	"github.com/google/uuid"
)

func TestAcceptArgumentsPreserveOwnerlessPlan(t *testing.T) {
	t.Parallel()
	plan := customerPlanFixture(t)
	request := provisioning.AcceptedRequest{
		ControlTenantID:      uuid.MustParse("11111111-1111-7111-8111-111111111111"),
		RequestedBySubjectID: uuid.MustParse("22222222-2222-7222-8222-222222222222"),
		Customer: provisioning.CustomerRequest{
			Name: "Acme", Slug: "acme",
			FirstAdmin: provisioning.FirstAdmin{
				DisplayName: "Ada", Email: "ada@example.com",
			},
		},
		JobID: uuid.MustParse("33333333-3333-7333-8333-333333333333"),
	}
	arguments := acceptArguments(plan, request)
	if got, want := len(arguments), 38; got != want {
		t.Fatalf("accept argument count = %d, want %d", got, want)
	}
	if arguments[14] != nil {
		t.Fatalf("instance owner binding = %v, want nil", arguments[14])
	}
	if arguments[26] != int64(1) {
		t.Fatalf("authority generation = %v, want 1", arguments[26])
	}
	if arguments[27] != nil {
		t.Fatalf("previous authority head = %v, want nil", arguments[27])
	}
}

func customerPlanFixture(t *testing.T) tenantapi.OnboardingPlan {
	t.Helper()
	tenantID := uuid.MustParse("44444444-4444-7444-8444-444444444444")
	schema, err := domain.SchemaNameFromUUID(tenantID.String())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	input := tenantapi.OnboardingPlanInput{
		OperationID:   uuid.MustParse("55555555-5555-7555-8555-555555555555"),
		EnvironmentID: uuid.MustParse("66666666-6666-7666-8666-666666666666"),
		TenantID:      tenantID, FirstUserID: uuid.MustParse("77777777-7777-7777-8777-777777777777"),
		RootOrganizationID:      uuid.MustParse("88888888-8888-7888-8888-888888888888"),
		ProtectedAdminBindingID: uuid.MustParse("99999999-9999-7999-8999-999999999999"),
		Kind:                    tenantapi.OnboardingSaaSCustomerProvision,
		DeploymentMode:          tenantapi.DeploymentModeSaaS,
		BindingPurpose:          tenantapi.TenantBindingSaaSCustomer,
		SchemaName:              schema.String(), IdempotencyKey: "customer-acme",
		RequestSHA256: strings.Repeat("1", 64), ManifestSHA256: strings.Repeat("2", 64),
		Offering: tenantapi.BaseOfferingPin{VersionID: uuid.MustParse("aaaaaaaa-aaaa-7aaa-8aaa-aaaaaaaaaaaa"),
			Revision: 1, SHA256: strings.Repeat("3", 64), EffectiveAt: now},
		Authority: tenantapi.ProtectedAuthorityPlan{
			RegistryBindingID: uuid.MustParse("bbbbbbbb-bbbb-7bbb-8bbb-bbbbbbbbbbbb"),
			ScopeKind:         tenantapi.AuthorityScopeTenant, ScopeID: tenantID, Generation: 1,
			StateSHA256: strings.Repeat("4", 64), ProposalSHA256: strings.Repeat("5", 64),
		},
		CreatedAt: now,
	}
	plan, err := tenantapi.NewOnboardingPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
