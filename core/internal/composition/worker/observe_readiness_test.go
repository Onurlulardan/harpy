package worker

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestWorkerReadinessPausesAndResumesClaimAdmission(t *testing.T) {
	now := time.Date(2026, 7, 22, 20, 0, 0, 0, time.UTC)
	gate := admission.NewGate(func() time.Time { return now })
	observe := observeWorkerReadiness(gate)
	observe(workerObservation(t, now, health.States{}))
	if gate.Accepting() {
		t.Fatal("failed readiness opened worker admission")
	}
	ready := health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}
	observe(workerObservation(t, now, ready))
	if !gate.Accepting() {
		t.Fatal("ready worker did not open claim admission")
	}
	observe(workerObservation(t, now, health.States{}))
	if gate.Accepting() {
		t.Fatal("dependency failure left claim admission open")
	}
}

func workerObservation(t *testing.T, now time.Time, states health.States) health.Observation {
	t.Helper()
	observation, err := health.NewObservation(states, now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return observation
}
