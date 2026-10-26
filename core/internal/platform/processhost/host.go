package processhost

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
)

type Dependencies struct {
	Monitor            *health.Monitor
	Poller             *health.Poller
	Management         *httpserver.Bound
	Public             *httpserver.Bound
	CloseAdmission     func()
	Workload           Workload
	CloseDatabase      func()
	ForceCloseDatabase func()
	ShutdownTelemetry  func(context.Context) error
	DrainTimeout       time.Duration
	ForceFlushTimeout  time.Duration
}

type Host struct {
	dependencies Dependencies
	pollerCancel context.CancelFunc
	pollerDone   <-chan struct{}
	closeOnce    sync.Once
	forceOnce    sync.Once
}

func New(dependencies Dependencies) (*Host, error) {
	if dependencies.Monitor == nil || dependencies.Poller == nil ||
		dependencies.Management == nil {
		return nil, fmt.Errorf("process health and management dependencies are required")
	}
	if dependencies.CloseAdmission == nil || dependencies.CloseDatabase == nil ||
		dependencies.ForceCloseDatabase == nil ||
		dependencies.ShutdownTelemetry == nil {
		return nil, fmt.Errorf("process lifecycle callbacks are required")
	}
	if dependencies.DrainTimeout <= 0 || dependencies.ForceFlushTimeout <= 0 ||
		dependencies.ForceFlushTimeout > dependencies.DrainTimeout {
		return nil, fmt.Errorf("process lifecycle timeouts are invalid")
	}
	if err := dependencies.Workload.validate(); err != nil {
		return nil, err
	}
	return &Host{dependencies: dependencies}, nil
}

func (host *Host) closeDatabase() {
	host.closeOnce.Do(host.dependencies.CloseDatabase)
}

func (host *Host) forceCloseDatabase() {
	host.forceOnce.Do(host.dependencies.ForceCloseDatabase)
}
