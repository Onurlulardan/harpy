package eventdelivery

import (
	"context"
	"math/rand/v2"
	"time"
)

const (
	reconnectMinimum = time.Second
	reconnectMaximum = 30 * time.Second
)

type runtimeJitter struct{}

func (runtimeJitter) Uint64N(upperBound uint64) uint64 {
	return rand.Uint64N(upperBound)
}

func reconnectDelay(cap time.Duration, source jitterSource) time.Duration {
	if cap <= reconnectMinimum {
		return reconnectMinimum
	}
	span := cap - reconnectMinimum
	return reconnectMinimum + time.Duration(source.Uint64N(uint64(span)+1))
}

func nextReconnectCap(current time.Duration) time.Duration {
	if current >= reconnectMaximum/2 {
		return reconnectMaximum
	}
	return current * 2
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
