package iamactorv1

import (
	"crypto/ed25519"
	"sync/atomic"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type Clock func() time.Time

type Consumer struct {
	verifier actorassertionv1.Verifier
	now      Clock
	closed   atomic.Bool
	valid    bool
}

func NewConsumer(
	publicKeys map[string]ed25519.PublicKey,
	now Clock,
) (*Consumer, error) {
	if now == nil {
		return nil, newError(CodeInvalid)
	}
	verifier, err := actorassertionv1.NewVerifier(publicKeys)
	if err != nil {
		return nil, newError(CodeInvalid)
	}
	return &Consumer{verifier: verifier, now: now, valid: true}, nil
}

func (consumer *Consumer) Ready() bool {
	return consumer != nil && consumer.valid && consumer.now != nil &&
		!consumer.closed.Load()
}

func (consumer *Consumer) Close() {
	if consumer != nil {
		consumer.closed.Store(true)
	}
}
