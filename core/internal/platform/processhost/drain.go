package processhost

import (
	"context"
	"errors"
)

func (host *Host) beginDrain() {
	host.dependencies.CloseAdmission()
	if host.dependencies.Workload.configured() {
		host.dependencies.Workload.BeginDrain()
	}
	host.dependencies.Monitor.BeginDrain()
	if host.pollerCancel != nil {
		host.pollerCancel()
	}
}

func (host *Host) waitRuntime(ctx context.Context) error {
	if err := host.waitPoller(ctx); err != nil {
		return err
	}
	var publicErr error
	if host.dependencies.Public != nil {
		publicErr = host.dependencies.Public.Shutdown(ctx)
	}
	var workloadErr error
	if host.dependencies.Workload.configured() {
		workloadErr = host.dependencies.Workload.Wait(ctx)
	}
	if err := errors.Join(publicErr, workloadErr); err != nil {
		return err
	}
	host.closeDatabase()
	return nil
}

func (host *Host) forceRuntime() error {
	if host.pollerCancel != nil {
		host.pollerCancel()
	}
	var publicErr error
	if host.dependencies.Public != nil {
		publicErr = host.dependencies.Public.Close()
	}
	var workloadErr error
	if host.dependencies.Workload.configured() {
		workloadErr = host.dependencies.Workload.Force()
	}
	host.forceCloseDatabase()
	return errors.Join(publicErr, workloadErr)
}

func (host *Host) forceHost() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), host.dependencies.ForceFlushTimeout,
	)
	defer cancel()
	telemetryErr := host.dependencies.ShutdownTelemetry(ctx)
	managementErr := host.dependencies.Management.Close()
	return errors.Join(telemetryErr, managementErr)
}
