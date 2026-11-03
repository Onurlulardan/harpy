package environmentoperator

import (
	"errors"
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestInitializeNamesRefreshMismatch(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.preserveOnboarding = true
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) {
		t.Fatalf("refresh mismatch error = %v", err)
	}
	diagnostic := Diagnose(err)
	if diagnostic.Stage != "tenant_migration.refresh_onboarding" ||
		diagnostic.ErrorCode != "onboarding_state_mismatch" ||
		diagnostic.Detail == "" {
		t.Fatalf("diagnostic = %+v error=%v", diagnostic, err)
	}
}

func TestInitializeNamesEnqueueFailure(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.enqueueErr = errors.New("tenant migration executor is not claim ready")
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) || !errors.Is(err, fixture.migration.enqueueErr) {
		t.Fatalf("enqueue error = %v", err)
	}
	if Diagnose(err).Stage != "tenant_migration.enqueue" {
		t.Fatalf("diagnostic = %+v", Diagnose(err))
	}
}

func TestInitializeNamesMigrationDeadline(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.states = []string{"pending"}
	fixture.migration.databaseStep = maximumMigrationWait
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	diagnostic := Diagnose(err)
	if !errors.Is(err, ErrResumable) || diagnostic.Stage != "tenant_migration.wait_deadline" ||
		diagnostic.ErrorCode != "migration_deadline" {
		t.Fatalf("deadline diagnostic = %+v error=%v", diagnostic, err)
	}
}
