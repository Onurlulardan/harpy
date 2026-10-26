package postgres

import (
	"context"
	"sync/atomic"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type ConsumerStore interface {
	BeginEvent(context.Context, eventsv1.Envelope) (EventDecision, error)
	CompleteDatabaseEvent(context.Context, EventEffectPermit, eventsv1.EffectReceipt) error
	CompleteExternalEvent(
		context.Context,
		EventEffectPermit,
		eventsv1.EffectReceipt,
		effectsv1.Receipt,
	) error
	BeginBootstrap(context.Context, eventsv1.BootstrapReceipt) (BootstrapDecision, error)
	CompleteBootstrap(context.Context, BootstrapPermit) error
}

type EventDecision interface{ eventDecision() }

type DuplicateEvent struct{ eventID string }

func (DuplicateEvent) eventDecision()        {}
func (value DuplicateEvent) EventID() string { return value.eventID }

type ParkedSequenceGap struct {
	eventID  string
	expected int64
	received int64
	rootID   string
	jobID    string
	created  bool
}

func (ParkedSequenceGap) eventDecision()                {}
func (value ParkedSequenceGap) EventID() string         { return value.eventID }
func (value ParkedSequenceGap) ExpectedSequence() int64 { return value.expected }
func (value ParkedSequenceGap) ReceivedSequence() int64 { return value.received }
func (value ParkedSequenceGap) RootID() string          { return value.rootID }
func (value ParkedSequenceGap) ReconciliationJobID() string {
	return value.jobID
}
func (value ParkedSequenceGap) Created() bool { return value.created }

type EventEffectPermit struct {
	owner    *consumerIdentity
	envelope eventsv1.Envelope
	policy   eventsv1.ConsumerPolicy
	used     *atomic.Bool
}

func (EventEffectPermit) eventDecision() {}

type BootstrapDecision interface{ bootstrapDecision() }

type DuplicateBootstrap struct{ receiptID string }

func (DuplicateBootstrap) bootstrapDecision()      {}
func (value DuplicateBootstrap) ReceiptID() string { return value.receiptID }

type BootstrapPermit struct {
	owner   *consumerIdentity
	receipt eventsv1.BootstrapReceipt
	digest  string
	used    *atomic.Bool
}

func (BootstrapPermit) bootstrapDecision() {}
