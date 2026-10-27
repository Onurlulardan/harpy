package eventdelivery

import (
	"context"
	"sync"
	"time"
)

type connectorFunction func(context.Context, Registration) (brokerSession, error)

func (function connectorFunction) Open(
	ctx context.Context,
	registration Registration,
) (brokerSession, error) {
	return function(ctx, registration)
}

type sessionFunctions struct {
	run   func(context.Context) error
	close func() error
}

func (session sessionFunctions) Run(ctx context.Context) error { return session.run(ctx) }
func (session sessionFunctions) Close() error                  { return session.close() }

type maximumJitter struct{}

func (maximumJitter) Uint64N(upperBound uint64) uint64 { return upperBound - 1 }

type recordedWait struct {
	mutex  sync.Mutex
	delays []time.Duration
}

func (wait *recordedWait) call(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	wait.mutex.Lock()
	wait.delays = append(wait.delays, delay)
	wait.mutex.Unlock()
	return nil
}

func (wait *recordedWait) values() []time.Duration {
	wait.mutex.Lock()
	defer wait.mutex.Unlock()
	return append([]time.Duration(nil), wait.delays...)
}

func waitSignal(ctx context.Context, signal <-chan string) (string, error) {
	select {
	case value := <-signal:
		return value, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
