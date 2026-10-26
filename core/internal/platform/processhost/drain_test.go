package processhost

import (
	"context"
	"reflect"
	"testing"
)

func TestWaitRuntimeClosesDatabaseAfterPollerAndWorkload(t *testing.T) {
	events := make([]string, 0, 2)
	pollerDone := make(chan struct{})
	close(pollerDone)
	host := &Host{
		dependencies: Dependencies{
			Workload: Workload{
				BeginDrain: func() {},
				Wait: func(context.Context) error {
					events = append(events, "workload")
					return nil
				},
				Force: func() error { return nil },
			},
			CloseDatabase: func() { events = append(events, "database") },
		},
		pollerDone: pollerDone,
	}
	if err := host.waitRuntime(context.Background()); err != nil {
		t.Fatalf("wait runtime: %v", err)
	}
	if !reflect.DeepEqual(events, []string{"workload", "database"}) {
		t.Fatalf("drain order=%v", events)
	}
}

func TestForceRuntimeClosesDatabaseOnlyOnce(t *testing.T) {
	closed := 0
	host := &Host{dependencies: Dependencies{
		ForceCloseDatabase: func() { closed++ },
	}}
	if err := host.forceRuntime(); err != nil {
		t.Fatalf("first force: %v", err)
	}
	if err := host.forceRuntime(); err != nil {
		t.Fatalf("second force: %v", err)
	}
	if closed != 1 {
		t.Fatalf("database close count=%d", closed)
	}
}
