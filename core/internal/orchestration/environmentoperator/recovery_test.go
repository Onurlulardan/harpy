package environmentoperator

import (
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestInitializeProvisionsTenantSchemaOnce(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	request := InitializeRequest{Manifest: manifest, RecoverySigning: recoveryFixture()}
	result, err := fixture.coordinator.Initialize(t.Context(), request)
	if err != nil || result.Outcome != OutcomeCompleted {
		t.Fatalf("initialize = %#v, %v", result, err)
	}
	if fixture.provision.calls != 1 || fixture.finalizer.calls != 1 {
		t.Fatalf("provision=%d finalizer=%d", fixture.provision.calls, fixture.finalizer.calls)
	}
}
