package dashboardexport

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (dispatcher *RenderDispatcher) run(scan, force context.Context) error {
	forceContext, forceCancel := context.WithCancel(force)
	defer forceCancel()
	group := newRenderDispatchGroup(forceCancel)
	reaper := time.NewTicker(renderDispatcherReapEvery)
	defer reaper.Stop()
	for {
		if !dispatcher.startupReaped && dispatcher.admission.Accepting() {
			if err := dispatcher.spools.Reap(scan, dispatcher.claims,
				dispatcher.clock.Now()); err != nil {
				forceCancel()
				group.wait.Wait()
				return fmt.Errorf("reap Analytics export spool at startup: %w", err)
			}
			dispatcher.startupReaped = true
		}
		for group.available() && scan.Err() == nil {
			worked, err := dispatcher.claimAndLaunch(scan, forceContext, group)
			if err != nil {
				forceCancel()
				group.wait.Wait()
				return err
			}
			if !worked {
				break
			}
		}
		select {
		case <-scan.Done():
			group.wait.Wait()
			return nil
		case err := <-group.failures:
			group.wait.Wait()
			return err
		case <-group.released:
		case <-reaper.C:
			if dispatcher.admission.Accepting() {
				if err := dispatcher.spools.Reap(forceContext, dispatcher.claims,
					dispatcher.clock.Now()); err != nil {
					forceCancel()
					group.wait.Wait()
					return fmt.Errorf("reap Analytics dashboard export spool: %w", err)
				}
			}
		case <-time.After(renderDispatcherIdlePoll):
		}
	}
}

func (dispatcher *RenderDispatcher) execute(
	ctx context.Context, dispatch renderDispatch,
) error {
	heartbeats := time.NewTicker(renderDispatcherHeartbeat)
	defer heartbeats.Stop()
	return dispatcher.executeWithHeartbeats(ctx, dispatch, heartbeats.C)
}

func (dispatcher *RenderDispatcher) executeWithHeartbeats(
	force context.Context, dispatch renderDispatch, heartbeats <-chan time.Time,
) error {
	process, cancel := context.WithCancelCause(force)
	defer cancel(nil)
	type execution struct {
		result RenderExecutionResult
		err    error
	}
	done := make(chan execution, 1)
	go func() {
		result, err := dispatcher.executor.ExecuteClaimedRender(
			process, dispatch.attempt, dispatch.spool)
		done <- execution{result: result, err: err}
	}()
	for {
		select {
		case output := <-done:
			return validateRenderExecution(output.result, output.err)
		case observedAt, open := <-heartbeats:
			if !open {
				cancel(context.Canceled)
				output := <-done
				return validateRenderExecution(output.result, output.err)
			}
			lease, err := dispatcher.claims.CheckpointRender(
				force, dispatch.attempt, observedAt)
			if errors.Is(err, ErrDispatchLeaseStale) {
				cancel(ErrDispatchLeaseStale)
				<-done
				return nil
			}
			if err != nil {
				cancel(err)
				output := <-done
				executeErr := validateRenderExecution(output.result, output.err)
				if executeErr != nil && !errors.Is(executeErr, context.Canceled) {
					return errors.Join(fmt.Errorf(
						"checkpoint Analytics dashboard export: %w", err), executeErr)
				}
				return fmt.Errorf("checkpoint Analytics dashboard export: %w", err)
			}
			if !validTime(lease.LeaseUntil) || !lease.LeaseUntil.After(observedAt) {
				cancel(ErrDispatchLeaseStale)
				<-done
				return fmt.Errorf("checkpoint Analytics dashboard export returned invalid lease")
			}
			if lease.CancellationRequested {
				cancel(ErrWorkerCancelled)
			}
		case <-force.Done():
			cancel(context.Canceled)
			output := <-done
			return validateRenderExecution(output.result, output.err)
		}
	}
}

func validateRenderExecution(result RenderExecutionResult, executeErr error) error {
	if result.State == RenderExecutionLeaseLost &&
		(errors.Is(executeErr, ErrDispatchLeaseStale) ||
			errors.Is(executeErr, context.Canceled)) {
		return nil
	}
	if executeErr != nil {
		return fmt.Errorf("execute claimed Analytics dashboard export: %w", executeErr)
	}
	if result.CleanupError != nil {
		return fmt.Errorf("clean Analytics dashboard export attempt: %w",
			result.CleanupError)
	}
	switch result.State {
	case RenderExecutionReady, RenderExecutionCancelled, RenderExecutionRequeued,
		RenderExecutionDead, RenderExecutionOutcomeUnknown:
		return nil
	default:
		return fmt.Errorf("Analytics dashboard export execution result is invalid")
	}
}
