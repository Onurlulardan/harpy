package application

import (
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type OperationsService struct {
	health platformopsapi.HealthReader
	traces ports.TraceStore
	alerts ports.AlertStore
}

func NewOperationsService(
	health platformopsapi.HealthReader,
	traces ports.TraceStore,
	alerts ports.AlertStore,
) *OperationsService {
	if health == nil || traces == nil || alerts == nil {
		return nil
	}
	return &OperationsService{health: health, traces: traces, alerts: alerts}
}
