package health

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestSnapshotReasonPrecedence(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	monitor, err := NewMonitor(runtimev1.ServiceCoreAPI, 30*time.Second, clock)
	if err != nil {
		t.Fatal(err)
	}
	monitor.Update(readyHealthStates())
	assertReason(t, monitor, runtimev1.ReadinessReasonStarting)
	monitor.MarkRunning()
	assertReason(t, monitor, runtimev1.ReadinessReasonNone)
	failed := readyHealthStates()
	failed.Database = false
	monitor.Update(failed)
	assertReason(t, monitor, runtimev1.ReadinessReasonDependencyFailed)
	monitor.BeginDrain()
	assertReason(t, monitor, runtimev1.ReadinessReasonDraining)
}

func TestSnapshotFailsClosedWhenStaleOrClockMovesBack(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	monitor, err := NewMonitor(runtimev1.ServiceCoreWorker, 30*time.Second, clock)
	if err != nil {
		t.Fatal(err)
	}
	monitor.Update(readyHealthStates())
	monitor.MarkRunning()
	now = now.Add(61 * time.Second)
	assertReason(t, monitor, runtimev1.ReadinessReasonDependencyFailed)
	now = now.Add(-2 * time.Minute)
	assertReason(t, monitor, runtimev1.ReadinessReasonDependencyFailed)
}

func readyHealthStates() States {
	return States{
		Database: true, DatabaseRole: true,
		DeploymentPolicy: true, SchemaContract: true,
	}
}

func TestNewMonitorRejectsInvalidInputs(t *testing.T) {
	if _, err := NewMonitor("unknown", time.Second, nil); err == nil {
		t.Fatal("expected invalid service rejection")
	}
	if _, err := NewMonitor(runtimev1.ServiceCoreAPI, 0, nil); err == nil {
		t.Fatal("expected invalid period rejection")
	}
}

func assertReason(t *testing.T, monitor *Monitor, expected runtimev1.ReadinessReason) {
	t.Helper()
	snapshot, err := monitor.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Reason != expected {
		t.Fatalf("expected reason %s, got %s", expected, snapshot.Reason)
	}
}
