package onboarding

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type installationTestRow struct{}

func (installationTestRow) Scan(destinations ...any) error {
	*destinations[0].(*bool) = false
	return nil
}

func TestInstallationBindingPassesOperationKindAsTenthArgument(t *testing.T) {
	var statement string
	var arguments []any
	binding, err := BindInstallation(InstallationDatabase{
		QueryRow: func(_ context.Context, query string, values ...any) InstallationRow {
			statement, arguments = query, values
			return installationTestRow{}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	input := InstallationTenant{
		OperationID: uuid.New(), EnvironmentID: uuid.New(), TenantID: uuid.New(),
		OwnerUserID: uuid.New(), Name: "Tenant", Slug: "tenant",
		SchemaName: "t_abcdefghijklmnopqrstuvwxyz", BindingPurpose: "saas_customer",
		ManifestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Kind:           "saas_customer_tenant_provision",
	}
	if _, err = binding.ActivateTenant(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if statement != activationInitializeSQL || len(arguments) != 10 ||
		arguments[9] != input.Kind {
		t.Fatalf("activation query kind argument=%v count=%d", arguments, len(arguments))
	}
}
