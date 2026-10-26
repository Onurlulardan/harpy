package deadletterdelivery

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type scriptedOpen struct {
	session Session
	err     error
}

type scriptedConnector struct {
	mutex   *sync.Mutex
	results []scriptedOpen
	opened  chan int
	count   int
}

func newScriptedConnector(results ...scriptedOpen) *scriptedConnector {
	return &scriptedConnector{
		mutex: &sync.Mutex{}, results: results, opened: make(chan int, len(results)+4),
	}
}

func (connector *scriptedConnector) Open(context.Context) (Session, error) {
	connector.mutex.Lock()
	index := connector.count
	connector.count++
	connector.mutex.Unlock()
	connector.opened <- index + 1
	if index >= len(connector.results) {
		return nil, ErrSessionUnavailable
	}
	result := connector.results[index]
	return result.session, result.err
}

func (connector *scriptedConnector) OpenCount() int {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.count
}

type blockingSession struct {
	started   chan struct{}
	startOnce *sync.Once
	closed    chan struct{}
	closeOnce *sync.Once
	closeCall atomic.Int32
	runErr    error
}

func newBlockingSession() *blockingSession {
	return &blockingSession{
		started: make(chan struct{}), startOnce: &sync.Once{},
		closed: make(chan struct{}), closeOnce: &sync.Once{},
	}
}

func (session *blockingSession) Run(ctx context.Context) error {
	session.startOnce.Do(func() { close(session.started) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-session.closed:
		return session.runErr
	}
}

func (session *blockingSession) Close() error {
	session.closeCall.Add(1)
	session.closeOnce.Do(func() { close(session.closed) })
	return nil
}

type fixedJitter uint64

func openDeadLetterAdmission() *admission.Gate {
	gate := admission.NewGate(nil)
	gate.Open()
	return gate
}

func (source fixedJitter) Uint64N(upperBound uint64) uint64 {
	value := uint64(source)
	if value >= upperBound {
		return upperBound - 1
	}
	return value
}
