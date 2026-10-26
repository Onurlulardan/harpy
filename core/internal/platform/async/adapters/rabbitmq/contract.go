package rabbitmq

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

const (
	FailurePublish    = "BROKER_PUBLISH_FAILED"
	FailureNack       = "BROKER_NACK"
	FailureUnroutable = "BROKER_UNROUTABLE"
	FailureClosed     = "BROKER_CHANNEL_CLOSED"
	FailureTimeout    = "BROKER_CONFIRM_TIMEOUT"
)

type PublishRequest struct {
	Envelope        eventsv1.Envelope
	LeaseGeneration int64
}

type PublishReceipt struct {
	ConfirmSHA256 string
}

type PublishFailure struct {
	code   string
	digest string
}

func (failure *PublishFailure) Error() string { return failure.code }
func (failure *PublishFailure) Code() string  { return failure.code }
func (failure *PublishFailure) SHA256() string {
	return failure.digest
}

func newFailure(code string, eventID string, leaseGeneration int64) *PublishFailure {
	document := fmt.Sprintf("rabbitmq-failure.v1\n%s\n%s\n%d", code, eventID, leaseGeneration)
	digest := sha256.Sum256([]byte(document))
	return &PublishFailure{code: code, digest: hex.EncodeToString(digest[:])}
}
