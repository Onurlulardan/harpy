package environmentoperator

import (
	"errors"
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestOnPremiseInitializeCompletesAndReplaysAsVerifiedNoop(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	request := InitializeRequest{Manifest: manifest, RecoverySigning: recoveryFixture()}
	first, err := fixture.coordinator.Initialize(t.Context(), request)
	if err != nil || first.Outcome != OutcomeCompleted || !first.IsValid() {
		t.Fatalf("first initialize = %#v, %v", first, err)
	}
	second, err := fixture.coordinator.Initialize(t.Context(), request)
	if err != nil || second.Outcome != OutcomeNoop || !second.IsValid() ||
		second.OperationID != first.OperationID || second.TenantID != first.TenantID {
		t.Fatalf("replay initialize = %#v, %v", second, err)
	}
	if fixture.migration.enqueues != 1 || fixture.migration.readyCalls != 1 ||
		fixture.finalizer.calls != 1 || fixture.onboarding.records != 0 ||
		fixture.authority.prepares != 0 || fixture.authority.acknowledgements != 0 {
		t.Fatalf("side effects migration=%d/%d finalizer=%d records=%d authority=%d/%d",
			fixture.migration.enqueues, fixture.migration.readyCalls,
			fixture.finalizer.calls, fixture.onboarding.records,
			fixture.authority.prepares, fixture.authority.acknowledgements)
	}
}

func TestSaaSInitializePinsAndAcknowledgesGenerationOneAuthority(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentSaaS)
	fixture := newOperatorFixture(t, manifest)
	result, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{Manifest: manifest})
	if err != nil || result.Outcome != OutcomeCompleted || !result.IsValid() {
		t.Fatalf("SaaS initialize = %#v, %v", result, err)
	}
	if fixture.authority.prepares != 1 || fixture.authority.acknowledgements != 1 ||
		fixture.finalizer.calls != 1 {
		t.Fatalf("authority prepare=%d ack=%d finalizer=%d",
			fixture.authority.prepares, fixture.authority.acknowledgements,
			fixture.finalizer.calls)
	}
}

func TestInitializeRejectsCrossModeRecoveryMaterial(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentSaaS)
	fixture := newOperatorFixture(t, manifest)
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-mode recovery error = %v", err)
	}
}
