package tenancy

import (
	"strings"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestActivateTenantInputRequiresExactKindPurposePair(t *testing.T) {
	id := func() tenantapi.UUID { return tenantapi.UUID(uuid.New()) }
	base := ActivateTenantInput{
		OperationID: id(), EnvironmentID: id(), TenantID: id(), OwnerUserID: id(),
		Name: "Tenant", Slug: "tenant", SchemaName: "t_abcdefghijklmnopqrstuvwxyz",
		ManifestSHA256: strings.Repeat("a", 64),
	}
	for _, test := range []struct {
		name, kind, purpose string
		valid               bool
	}{
		{"environment onprem", "environment_initialize", "onpremise_primary", true},
		{"environment saas", "environment_initialize", "saas_control", true},
		{"customer", "saas_customer_tenant_provision", "saas_customer", true},
		{"customer under environment", "environment_initialize", "saas_customer", false},
		{"environment under customer", "saas_customer_tenant_provision", "saas_control", false},
		{"unknown kind", "unknown", "saas_customer", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := base
			input.Kind, input.BindingPurpose = test.kind, test.purpose
			if got := validActivateTenantInput(input); got != test.valid {
				t.Fatalf("validActivateTenantInput()=%v want %v", got, test.valid)
			}
		})
	}
}
