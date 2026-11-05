package environmentoperator

import (
	"errors"
	"strings"
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/google/uuid"
)

func TestSchemaNameUsesPrimaryTenantID(t *testing.T) {
	tenantID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	name, err := schemaNameForTenant(tenantID)
	if err != nil || name != "tenant_00112233445566778899aabbccddeeff" {
		t.Fatalf("name=%s err=%v", name, err)
	}
	if _, err = schemaNameForTenant(uuid.Nil); !errors.Is(err, ErrInvalid) ||
		!strings.Contains(err.Error(), "<empty>") {
		t.Fatal(err)
	}
}

func TestInitializeWaitsBeforeAcceptWhenSchemaIsMissing(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.provision.schemaErr = ErrTenantProvisioningPending
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) || !errors.Is(err, ErrTenantProvisioningPending) {
		t.Fatal(err)
	}
	if Diagnose(err).Stage != "tenant_schema.provision" ||
		Diagnose(err).ErrorCode != "provision_failed" ||
		fixture.onboarding.accepts != 0 || fixture.finalizer.calls != 0 {
		t.Fatalf("diagnostic=%+v accepts=%d finalizer=%d",
			Diagnose(err), fixture.onboarding.accepts, fixture.finalizer.calls)
	}
}

func TestInvalidSchemaNameIsNotResumable(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.provision.schemaErr = invalidTenantSchema("harpy")
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrInvalid) || errors.Is(err, ErrResumable) ||
		!strings.Contains(err.Error(), "tenant schema name is invalid: harpy") ||
		fixture.onboarding.accepts != 0 {
		t.Fatalf("err=%v accepts=%d", err, fixture.onboarding.accepts)
	}
}
