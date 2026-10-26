package platformops

import (
	"context"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

type runtimeHealthReader struct {
	monitor *health.Monitor
	clock   runtimeClock
}

func (reader *runtimeHealthReader) ReadSystemHealth(
	ctx context.Context,
) (platformopsapi.SystemHealth, error) {
	if reader == nil || reader.monitor == nil || ctx == nil || ctx.Err() != nil {
		return platformopsapi.SystemHealth{}, platformopsapi.ErrUnavailable
	}
	snapshot, err := reader.monitor.Snapshot()
	if err != nil {
		return platformopsapi.SystemHealth{}, platformopsapi.ErrUnavailable
	}
	now := reader.clock.Now()
	serviceStatus := platformopsapi.HealthUnhealthy
	if snapshot.Status == runtimev1.ReadinessStatusReady {
		serviceStatus = platformopsapi.HealthHealthy
	}
	result := platformopsapi.SystemHealth{Status: serviceStatus, ObservedAt: now,
		Services: []platformopsapi.HealthComponent{{Name: string(snapshot.Service),
			Status: serviceStatus, LastChecked: now}},
		Infrastructure: make([]platformopsapi.HealthComponent, len(snapshot.Checks))}
	for index, check := range snapshot.Checks {
		status := platformopsapi.HealthUnhealthy
		if check.Status == runtimev1.ReadinessStatusReady {
			status = platformopsapi.HealthHealthy
		}
		result.Infrastructure[index] = platformopsapi.HealthComponent{
			Name: string(check.Name), Status: status, LastChecked: now}
	}
	return result, nil
}

var _ platformopsapi.HealthReader = (*runtimeHealthReader)(nil)
