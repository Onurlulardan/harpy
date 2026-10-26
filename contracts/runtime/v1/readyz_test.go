package runtimev1

import (
	"encoding/json"
	"testing"
)

func TestNewReadyzMarshalsCanonicalReadyWire(t *testing.T) {
	t.Parallel()
	readyz, err := NewReadyz(ServiceCoreWorker, ReadinessReasonNone, allReadyStates())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(readyz)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"status":"ready","service":"core-worker","reason":"none","checks":[` +
		`{"name":"database","status":"ready"},` +
		`{"name":"database_role","status":"ready"},` +
		`{"name":"deployment_policy","status":"ready"},` +
		`{"name":"schema_contract","status":"ready"}]}`
	if string(encoded) != want {
		t.Fatalf("unexpected wire\nwant: %s\n got: %s", want, encoded)
	}
}

func TestNewReadyzMarshalsDependencyFailure(t *testing.T) {
	t.Parallel()
	states := allReadyStates()
	states.DatabaseRole = ReadinessStatusNotReady
	readyz, err := NewReadyz(ServiceCoreAPI, ReadinessReasonDependencyFailed, states)
	if err != nil {
		t.Fatal(err)
	}
	if readyz.Status != ReadinessStatusNotReady {
		t.Fatalf("unexpected status %q", readyz.Status)
	}
	if readyz.Checks[1].Name != ReadinessCheckDatabaseRole {
		t.Fatalf("unexpected fixed check order: %#v", readyz.Checks)
	}
}

func TestNewReadyzAllowsLifecycleGateWithReadyDependencies(t *testing.T) {
	t.Parallel()
	for _, reason := range []ReadinessReason{
		ReadinessReasonStarting,
		ReadinessReasonDraining,
	} {
		readyz, err := NewReadyz(ServiceCoreMigrator, reason, allReadyStates())
		if err != nil {
			t.Fatalf("reason %q: %v", reason, err)
		}
		if readyz.Status != ReadinessStatusNotReady {
			t.Fatalf("reason %q unexpectedly ready", reason)
		}
	}
}

func TestNewReadyzRejectsIncoherentState(t *testing.T) {
	t.Parallel()
	failedStates := allReadyStates()
	failedStates.Database = ReadinessStatusNotReady
	tests := []struct {
		name    string
		service Service
		reason  ReadinessReason
		states  ReadinessStates
	}{
		{name: "service", service: Service("api"), reason: ReadinessReasonNone, states: allReadyStates()},
		{name: "reason", service: ServiceCoreAPI, reason: ReadinessReason("unknown"), states: allReadyStates()},
		{name: "none with failure", service: ServiceCoreAPI, reason: ReadinessReasonNone, states: failedStates},
		{name: "dependency failure without failure", service: ServiceCoreAPI, reason: ReadinessReasonDependencyFailed, states: allReadyStates()},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewReadyz(test.service, test.reason, test.states); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestReadyzMarshalRejectsWrongCheckOrder(t *testing.T) {
	t.Parallel()
	readyz, err := NewReadyz(ServiceCoreAPI, ReadinessReasonNone, allReadyStates())
	if err != nil {
		t.Fatal(err)
	}
	readyz.Checks[0], readyz.Checks[1] = readyz.Checks[1], readyz.Checks[0]
	if _, err := json.Marshal(readyz); err == nil {
		t.Fatal("invalid readiness crossed the JSON boundary")
	}
}

func allReadyStates() ReadinessStates {
	return ReadinessStates{
		Database:         ReadinessStatusReady,
		DatabaseRole:     ReadinessStatusReady,
		DeploymentPolicy: ReadinessStatusReady,
		SchemaContract:   ReadinessStatusReady,
	}
}
