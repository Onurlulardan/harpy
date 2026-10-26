package processhost

import (
	"context"
	"fmt"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
)

func (host *Host) Run(ctx context.Context) lifecycle.Outcome {
	signals := lifecycle.NotifySignals()
	defer signals.Stop()
	return host.run(ctx, signals.Channel())
}

func (host *Host) RunProcess(process *lifecycle.Process) lifecycle.Outcome {
	if process == nil || process.RuntimeContext() == nil {
		panic("process lifecycle is required")
	}
	return host.run(process.RuntimeContext(), process.Interrupts())
}

func (host *Host) run(
	ctx context.Context,
	interrupts <-chan os.Signal,
) lifecycle.Outcome {
	runtimeFailures := make(chan error, 3)
	host.startServers(runtimeFailures)
	host.startWorkloadMonitor(runtimeFailures)
	host.startPoller(ctx)
	hooks := lifecycle.NewHooks()
	if err := hooks.Register(lifecycle.Hook{
		Name: "runtime", Begin: func() error { return nil },
		Wait: host.waitRuntime, Force: host.forceRuntime,
	}); err != nil {
		panic(fmt.Sprintf("register process lifecycle: %v", err))
	}
	coordinator, err := lifecycle.NewCoordinator(lifecycle.Plan{
		BeginDrain: host.beginDrain, InFlight: hooks,
		FlushTelemetry: host.dependencies.ShutdownTelemetry,
		StopManagement: host.dependencies.Management.Shutdown,
		ForceClose:     host.forceHost, AbsoluteDeadline: host.dependencies.DrainTimeout,
	})
	if err != nil {
		panic(fmt.Sprintf("build process lifecycle: %v", err))
	}
	return coordinator.Run(ctx, interrupts, runtimeFailures)
}

func (host *Host) startServers(failures chan<- error) {
	go reportServeFailure(host.dependencies.Management, failures)
	if host.dependencies.Public != nil {
		go reportServeFailure(host.dependencies.Public, failures)
	}
}

func reportServeFailure(server *httpserver.Bound, failures chan<- error) {
	if err := server.Serve(); err != nil {
		failures <- err
	}
}

func (host *Host) startWorkloadMonitor(failures chan<- error) {
	if !host.dependencies.Workload.configured() {
		return
	}
	go func() {
		if err, open := <-host.dependencies.Workload.Failures; open && err != nil {
			failures <- err
		}
	}()
}
