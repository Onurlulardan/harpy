package deadletter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

const (
	EventPermanentCode = "EVENT_PERMANENT_REJECTION"
	EventVersionCode   = "EVENT_VERSION_UNKNOWN"
	EventDeliveryCode  = "EVENT_DELIVERY_LIMIT"
	EventSequenceCode  = "EVENT_SEQUENCE_GAP"
)

type DeliveryEvidence struct {
	Reason        ReasonClass
	DeliveryCount int
}

type EventRoot struct {
	rootID           uuid.UUID
	tenantID         *uuid.UUID
	scopeID          uuid.UUID
	consumer         eventsv1.ConsumerPolicy
	envelope         eventsv1.Envelope
	sourceSHA256     string
	reasonClass      ReasonClass
	reasonCode       string
	errorSHA256      string
	payload          []byte
	redactedEnvelope []byte
}

func NewEventRoot(
	document []byte,
	queueName string,
	evidence DeliveryEvidence,
	consumers eventsv1.ConsumerCatalog,
) (EventRoot, error) {
	envelope, err := eventsv1.DecodeDeadLetterEnvelope(document)
	if err != nil {
		return EventRoot{}, fmt.Errorf("decode broker dead-letter: %w", err)
	}
	consumer, err := consumers.ResolveQueue(queueName)
	if err != nil {
		return EventRoot{}, fmt.Errorf("resolve broker dead-letter queue: %w", err)
	}
	reasonClass, reasonCode, err := classifyEventReason(envelope, consumer, evidence, consumers)
	if err != nil {
		return EventRoot{}, err
	}
	canonical, err := canonicaljsonv1.Canonicalize(document, eventsv1.MaximumDocumentBytes)
	if err != nil {
		return EventRoot{}, fmt.Errorf("canonicalize broker dead-letter: %w", err)
	}
	redacted, err := marshalEventMetadata(envelope)
	if err != nil {
		return EventRoot{}, err
	}
	root, err := makeEventRoot(envelope, consumer, canonical, redacted, reasonClass, reasonCode, evidence)
	if err != nil {
		return EventRoot{}, err
	}
	return root, nil
}

func makeEventRoot(
	envelope eventsv1.Envelope,
	consumer eventsv1.ConsumerPolicy,
	canonical []byte,
	redacted []byte,
	reasonClass ReasonClass,
	reasonCode string,
	evidence DeliveryEvidence,
) (EventRoot, error) {
	scopeID, err := uuid.Parse(envelope.ProducerAuthority.ScopeID)
	if err != nil || scopeID == uuid.Nil {
		return EventRoot{}, fmt.Errorf("broker dead-letter scope is invalid")
	}
	var tenantID *uuid.UUID
	if envelope.TenantID != nil {
		parsed, parseErr := uuid.Parse(*envelope.TenantID)
		if parseErr != nil || parsed == uuid.Nil {
			return EventRoot{}, fmt.Errorf("broker dead-letter tenant is invalid")
		}
		tenantID = &parsed
	}
	sourceDigest := sha256.Sum256(canonical)
	sourceSHA := hex.EncodeToString(sourceDigest[:])
	return EventRoot{
		rootID:   eventRootID(envelope.EventID, consumer, sourceSHA, reasonClass),
		tenantID: tenantID, scopeID: scopeID, consumer: consumer, envelope: envelope,
		sourceSHA256: sourceSHA, reasonClass: reasonClass, reasonCode: reasonCode,
		errorSHA256: eventErrorDigest(envelope.EventID, consumer, reasonCode, evidence),
		payload:     envelope.PayloadClone(), redactedEnvelope: append([]byte(nil), redacted...),
	}, nil
}

func (value EventRoot) RootID() uuid.UUID                 { return value.rootID }
func (value EventRoot) ScopeID() uuid.UUID                { return value.scopeID }
func (value EventRoot) Consumer() eventsv1.ConsumerPolicy { return value.consumer }
func (value EventRoot) Envelope() eventsv1.Envelope       { return cloneEnvelope(value.envelope) }
func (value EventRoot) SourceSHA256() string              { return value.sourceSHA256 }
func (value EventRoot) ReasonClass() ReasonClass          { return value.reasonClass }
func (value EventRoot) ReasonCode() string                { return value.reasonCode }
func (value EventRoot) ErrorSHA256() string               { return value.errorSHA256 }
func (value EventRoot) Payload() []byte                   { return append([]byte(nil), value.payload...) }
func (value EventRoot) RedactedEnvelope() []byte {
	return append([]byte(nil), value.redactedEnvelope...)
}
func (value EventRoot) TenantID() *uuid.UUID {
	if value.tenantID == nil {
		return nil
	}
	copy := *value.tenantID
	return &copy
}

func cloneEnvelope(value eventsv1.Envelope) eventsv1.Envelope {
	copy := value
	copy.Payload = value.PayloadClone()
	if value.TenantID != nil {
		tenant := *value.TenantID
		copy.TenantID = &tenant
	}
	return copy
}
