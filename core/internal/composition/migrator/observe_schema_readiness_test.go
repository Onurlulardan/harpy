package migrator

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestObserveSchemaReadinessCannotReopenAfterDrain(t *testing.T) {
	gate := &admission.Gate{}
	observe := observeSchemaReadiness(gate)
	ready := health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}
	observe(schemaObservation(t, ready))
	if !gate.Accepting() {
		t.Fatal("ready executor did not open admission")
	}
	observe(schemaObservation(t, health.States{}))
	if gate.Accepting() {
		t.Fatal("unready executor kept admission open")
	}
	gate.BeginDrain()
	observe(schemaObservation(t, ready))
	if !gate.Draining() || gate.Accepting() {
		t.Fatal("readiness reopened a draining executor")
	}
}

func schemaObservation(t *testing.T, states health.States) health.Observation {
	t.Helper()
	observation, err := health.NewObservation(states, time.Now(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return observation
}
