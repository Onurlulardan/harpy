package outboxdelivery

import (
	"testing"
	"time"
)

type fixedJitter uint64

func (source fixedJitter) Uint64N(upperBound uint64) uint64 {
	value := uint64(source)
	if value >= upperBound {
		return upperBound - 1
	}
	return value
}

func TestReconnectDelayStaysInsideFinalBounds(t *testing.T) {
	settings := Settings{
		ConfirmTimeout: time.Second, ReconnectMinimum: time.Second,
		ReconnectMaximum: 30 * time.Second,
	}
	if got := reconnectDelay(time.Second, settings, fixedJitter(99)); got != time.Second {
		t.Fatalf("minimum reconnect delay = %s", got)
	}
	if got := reconnectDelay(8*time.Second, settings, fixedJitter(^uint64(0))); got != 8*time.Second {
		t.Fatalf("maximum sampled reconnect delay = %s", got)
	}
	if got := nextReconnectCap(16*time.Second, 30*time.Second); got != 30*time.Second {
		t.Fatalf("capped reconnect window = %s", got)
	}
}
