package api

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestObserveReadinessControlsAdmissionFromExactChecks(t *testing.T) {
	admission := &httpedge.Admission{}
	observe := observeReadiness(admission)

	observe(apiObservation(t, health.States{Database: true, DatabaseRole: true}))
	if admission.Accepting() {
		t.Fatal("partial readiness opened admission")
	}
	observe(apiObservation(t, health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}))
	if !admission.Accepting() {
		t.Fatal("complete readiness did not open admission")
	}
	observe(apiObservation(t, health.States{
		Database: false, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}))
	if admission.Accepting() {
		t.Fatal("dependency failure left admission open")
	}
	observe(apiObservation(t, health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}))
	if !admission.Accepting() {
		t.Fatal("dependency recovery did not reopen admission")
	}
}

func TestObserveReadinessCannotReopenDrainingAdmission(t *testing.T) {
	admission := &httpedge.Admission{}
	admission.BeginDrain()
	observeReadiness(admission)(apiObservation(t, health.States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}))
	if admission.Accepting() || !admission.Draining() {
		t.Fatal("readiness reopened draining admission")
	}
}

func apiObservation(t *testing.T, states health.States) health.Observation {
	t.Helper()
	observation, err := health.NewObservation(states, time.Now(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return observation
}
