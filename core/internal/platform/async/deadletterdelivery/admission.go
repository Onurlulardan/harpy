package deadletterdelivery

import (
	"context"
	"sync"
	"time"
)

const admissionPollInterval = 100 * time.Millisecond

func (runner *Runner) waitForAdmission(ctx context.Context) bool {
	for !runner.admission.Accepting() {
		timer := time.NewTimer(admissionPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
	}
	return true
}

func (runner *Runner) runAdmittedSession(
	ctx context.Context,
	session Session,
) (error, bool) {
	controlled := &admittedSession{Session: session}
	runner.setActive(controlled)
	defer func() {
		active := runner.detachActive()
		if active != nil {
			_ = closeSession(active)
		}
		_ = controlled.Close()
	}()
	sessionContext, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	paused := make(chan struct{}, 1)
	go runner.monitorAdmission(
		sessionContext, cancel, controlled, done, paused,
	)
	err := controlled.Run(sessionContext)
	close(done)
	select {
	case <-paused:
		return err, true
	default:
		return err, false
	}
}

func (runner *Runner) monitorAdmission(
	ctx context.Context,
	cancel context.CancelFunc,
	session *admittedSession,
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

type admittedSession struct {
	Session
	once sync.Once
	err  error
}

func (session *admittedSession) Close() error {
	session.once.Do(func() { session.err = session.Session.Close() })
	return session.err
}
