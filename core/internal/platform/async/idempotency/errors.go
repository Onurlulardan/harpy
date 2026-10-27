package idempotency

import "errors"

var (
	ErrInvalidIdentity = errors.New("http idempotency identity is invalid")
	ErrInvalidRequest  = errors.New("http idempotency request is invalid")
	ErrInvalidResponse = errors.New("http idempotency response is invalid")
	ErrInvalidReceipt  = errors.New("http idempotency receipt is invalid")
	ErrInProgress      = errors.New("IDEMPOTENCY_IN_PROGRESS")
	ErrKeyReused       = errors.New("IDEMPOTENCY_KEY_REUSED")
)
