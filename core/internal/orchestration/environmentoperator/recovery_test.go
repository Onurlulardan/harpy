package environmentoperator

import (
	"errors"
	"testing"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestInitializeWaitsUntilClaimReadyBecomesAvailable(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.ready = false
	fixture.migration.readyAfter = 3
	request := InitializeRequest{Manifest: manifest, RecoverySigning: recoveryFixture()}
	result, err := fixture.coordinator.Initialize(t.Context(), request)
	if err != nil || result.Outcome != OutcomeCompleted {
		t.Fatalf("waited initialize = %#v, %v", result, err)
	}
	if fixture.migration.readyCalls != 3 || fixture.migration.enqueues != 1 ||
		fixture.finalizer.calls != 1 {
		t.Fatalf("claim-ready calls=%d enqueues=%d finalizer=%d",
			fixture.migration.readyCalls, fixture.migration.enqueues,
			fixture.finalizer.calls)
	}
}

func TestInitializeTimesOutBeforeWorkWithoutClaimReady(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.ready = false
	fixture.coordinator.poll = time.Second
	request := InitializeRequest{Manifest: manifest, RecoverySigning: recoveryFixture()}
	if _, err := fixture.coordinator.Initialize(t.Context(), request); !errors.Is(
		err, ErrDependencyMismatch,
	) {
		t.Fatalf("claim-ready timeout error = %v", err)
	}
	if fixture.migration.enqueues != 0 || fixture.finalizer.calls != 0 {
		t.Fatal("work escaped claim-ready timeout fence")
	}
}

func TestInitializeUsesDurableFiveMinuteMigrationDeadline(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.states = []string{"pending"}
	fixture.migration.databaseStep = maximumMigrationWait
	_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
		Manifest: manifest, RecoverySigning: recoveryFixture(),
	})
	if !errors.Is(err, ErrResumable) || fixture.finalizer.calls != 0 {
		t.Fatalf("deadline result error=%v finalizer=%d", err, fixture.finalizer.calls)
	}
	if maximumMigrationWait != 5*time.Minute {
		t.Fatalf("migration wait = %s", maximumMigrationWait)
	}
}

func TestInitializeResumesAfterInterruptedMigrationWait(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	fixture := newOperatorFixture(t, manifest)
	fixture.migration.states = []string{"pending"}
	fixture.migration.databaseStep = time.Second
	*fixture.sleepErr = errors.New("operator interrupted")
	request := InitializeRequest{Manifest: manifest, RecoverySigning: recoveryFixture()}
	if _, err := fixture.coordinator.Initialize(t.Context(), request); !errors.Is(err, ErrResumable) {
		t.Fatalf("interrupted wait error = %v", err)
	}
	fixture.onboarding.activeAttempt = [16]byte{}
	fixture.migration.states, fixture.migration.index = []string{"succeeded"}, 0
	fixture.migration.databaseStep = time.Second
	*fixture.sleepErr = nil
	result, err := fixture.coordinator.Initialize(t.Context(), request)
	if err != nil || result.Outcome != OutcomeCompleted || fixture.finalizer.calls != 1 {
		t.Fatalf("resume after interrupt = %#v error=%v finalizer=%d",
			result, err, fixture.finalizer.calls)
	}
}

func TestInitializeClassifiesTerminalMigrationReceipts(t *testing.T) {
	for _, candidate := range []struct {
		state string
		want  error
	}{{"failed", ErrFailed}, {"fenced", ErrFenced},
		{"reconcile_required", ErrFenced}, {"cleanup_required", ErrFenced}} {
		t.Run(candidate.state, func(t *testing.T) {
			manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
			fixture := newOperatorFixture(t, manifest)
			fixture.migration.states = []string{candidate.state}
			_, err := fixture.coordinator.Initialize(t.Context(), InitializeRequest{
				Manifest: manifest, RecoverySigning: recoveryFixture(),
			})
			if !errors.Is(err, candidate.want) {
				t.Fatalf("state %s error = %v", candidate.state, err)
			}
		})
	}
}
