package eventdelivery

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const admissionPollInterval = 100 * time.Millisecond

func (runner *Runner) runRegistration(
	ctx context.Context,
	registration Registration,
) error {
	cap := reconnectMinimum
	key := registrationKey(registration.consumer, registration.version)
	for {
		if ctx.Err() != nil {
			return nil
		}
		if !runner.waitForAdmission(ctx) {
			return nil
		}
		session, err := runner.connector.Open(ctx, registration)
		if err == nil && session == nil {
			return fmt.Errorf(
				"consumer %s version %d: %w",
				registration.consumer, registration.version, ErrSessionUnavailable,
			)
		}
		if err != nil {
			if runner.waitReconnect(ctx, cap) != nil {
				return nil
			}
			cap = nextReconnectCap(cap)
			continue
		}
		cap = reconnectMinimum
		var paused bool
		err, paused = runner.runAdmittedSession(ctx, key, session)
		if ctx.Err() != nil {
			return nil
		}
		if paused {
			cap = reconnectMinimum
			continue
		}
		if runner.waitReconnect(ctx, cap) != nil {
			return nil
		}
		cap = nextReconnectCap(cap)
	}
}

func (runner *Runner) waitForAdmission(ctx context.Context) bool {
	for !runner.admission.Accepting() {
		if runner.wait(ctx, admissionPollInterval) != nil {
			return false
		}
	}
	return true
}

func (runner *Runner) runAdmittedSession(
	ctx context.Context, key string, session brokerSession,
) (error, bool) {
	controlled := &controlledSession{brokerSession: session}
	sessionContext, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	paused := make(chan struct{}, 1)
	go runner.monitorSessionAdmission(sessionContext, cancel, controlled, done, paused)
	err := runner.runSession(sessionContext, key, controlled)
	close(done)
	select {
	case <-paused:
		return err, true
	default:
		return err, false
	}
}

func (runner *Runner) monitorSessionAdmission(
	ctx context.Context,
	cancel context.CancelFunc,
	session *controlledSession,
	done <-chan struct{},
	paused chan<- struct{},
) {
	ticker := time.NewTicker(admissionPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if runner.admission.Accepting() {
				continue
			}
			select {
			case paused <- struct{}{}:
			default:
			}
			cancel()
			_ = session.Close()
			return
		}
	}
}

type controlledSession struct {
	brokerSession
	once sync.Once
	err  error
}

func (session *controlledSession) Close() error {
	session.once.Do(func() { session.err = session.brokerSession.Close() })
	return session.err
}

func (runner *Runner) runSession(
	ctx context.Context,
	key string,
	session brokerSession,
) error {
	runner.setActive(key, session)
	defer runner.setActive(key, nil)
	defer session.Close()
	return session.Run(ctx)
}

func (runner *Runner) waitReconnect(ctx context.Context, cap time.Duration) error {
	return runner.wait(ctx, reconnectDelay(cap, runner.jitter))
}
