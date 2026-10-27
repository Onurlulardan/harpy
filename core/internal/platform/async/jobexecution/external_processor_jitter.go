package jobexecution

import (
	"crypto/rand"
	"math/big"
)

type cryptographicJitter struct{}

func (cryptographicJitter) Uint64N(upperBound uint64) uint64 {
	if upperBound == 0 {
		panic("external job retry jitter bound is zero")
	}
	value, err := rand.Int(rand.Reader, new(big.Int).SetUint64(upperBound))
	if err != nil {
		panic("external job retry entropy is unavailable")
	}
	return value.Uint64()
}
