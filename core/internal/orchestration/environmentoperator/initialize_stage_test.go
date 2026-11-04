package environmentoperator

import (
	"errors"
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestInitializeNamesRefreshMismatch(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.provision.preserveOnboarding = true
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) {
		t.Fatalf("refresh mismatch error = %v", err)
	}
	diagnostic := Diagnose(err)
	if diagnostic.Stage != "tenant_schema.refresh_onboarding" ||
		diagnostic.ErrorCode != "onboarding_state_mismatch" ||
		diagnostic.Detail == "" {
		t.Fatalf("diagnostic = %+v error=%v", diagnostic, err)
	}
}

func TestInitializeNamesProvisionFailure(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.provision.err = errors.New("tenant provisioning container failed")
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) || !errors.Is(err, fixture.provision.err) {
		t.Fatalf("provision error = %v", err)
	}
	if Diagnose(err).Stage != "tenant_schema.provision" {
		t.Fatalf("diagnostic = %+v", Diagnose(err))
	}
}
