package outboxdelivery

import (
	"math/rand/v2"
	"time"
)

type jitterSource interface {
	Uint64N(uint64) uint64
}

type runtimeJitter struct{}

func (runtimeJitter) Uint64N(upperBound uint64) uint64 {
	return rand.Uint64N(upperBound)
}

func reconnectDelay(cap time.Duration, settings Settings, source jitterSource) time.Duration {
	if cap <= settings.ReconnectMinimum {
		return settings.ReconnectMinimum
	}
	span := cap - settings.ReconnectMinimum
	return settings.ReconnectMinimum +
		time.Duration(source.Uint64N(uint64(span)+1))
}

func nextReconnectCap(current time.Duration, maximum time.Duration) time.Duration {
	if current >= maximum/2 {
		return maximum
	}
	return current * 2
}
