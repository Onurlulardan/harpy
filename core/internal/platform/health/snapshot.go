package health

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func (monitor *Monitor) Snapshot() (runtimev1.Readyz, error) {
	monitor.mu.RLock()
	states := monitor.states
	observedAt := monitor.observedAt
	running := monitor.running
	draining := monitor.draining
	monitor.mu.RUnlock()

	if staleSnapshot(monitor.now(), observedAt, monitor.healthPeriod) {
		states = States{}
	}
	reason := readinessReason(states, running, draining)
	readyz, err := runtimev1.NewReadyz(
		monitor.service,
		reason,
		runtimev1.ReadinessStates{
			Database:         readinessStatus(states.Database),
			DatabaseRole:     readinessStatus(states.DatabaseRole),
			DeploymentPolicy: readinessStatus(states.DeploymentPolicy),
			SchemaContract:   readinessStatus(states.SchemaContract),
		},
	)
	if err != nil {
		return runtimev1.Readyz{}, fmt.Errorf("build readiness snapshot: %w", err)
	}
	return readyz, nil
}

func staleSnapshot(now, observedAt time.Time, healthPeriod time.Duration) bool {
	if observedAt.IsZero() {
		return true
	}
	age := now.Sub(observedAt)
	return age < 0 || age > 2*healthPeriod
}

func readinessReason(states States, running, draining bool) runtimev1.ReadinessReason {
	if draining {
		return runtimev1.ReadinessReasonDraining
	}
	if !running {
		return runtimev1.ReadinessReasonStarting
	}
	if !states.Database || !states.DatabaseRole || !states.DeploymentPolicy ||
		!states.SchemaContract {
		return runtimev1.ReadinessReasonDependencyFailed
	}
	return runtimev1.ReadinessReasonNone
}

func readinessStatus(ready bool) runtimev1.ReadinessStatus {
	if ready {
		return runtimev1.ReadinessStatusReady
	}
	return runtimev1.ReadinessStatusNotReady
}
