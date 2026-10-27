package deadletter

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	PermissionRead  = "async.deadletter.read"
	PermissionRetry = "async.deadletter.retry"

	MaximumPageSize    = 100
	MaximumTenantScope = 100
	MaximumReasonBytes = 2_048
	MaximumTicketBytes = 256
)

type SourceKind string

const (
	SourceJob   SourceKind = "job"
	SourceEvent SourceKind = "event"
)

type ReasonClass string

const (
	ReasonPermanent      ReasonClass = "permanent"
	ReasonUnknownVersion ReasonClass = "unknown_version"
	ReasonDeliveryLimit  ReasonClass = "delivery_limit"
	ReasonSequenceGap    ReasonClass = "sequence_gap"
)

type Cursor struct {
	RecordedAt time.Time
	RootID     uuid.UUID
}

type Entry struct {
	RootID            uuid.UUID
	TenantID          uuid.UUID
	SourceKind        SourceKind
	SourceID          uuid.UUID
	Consumer          string
	ConsumerVersion   int
	SourceContract    string
	SourceVersion     int
	ReasonClass       ReasonClass
	ReasonCode        string
	PayloadAvailable  bool
	RedactedEnvelope  json.RawMessage
	RecordedAt        time.Time
	PayloadExpiresAt  time.Time
	MetadataExpiresAt time.Time
	RetryActionCount  int
}

type Page struct {
	Entries []Entry
	Next    *Cursor
}

type RetryResult struct {
	ActionID      uuid.UUID
	SourceKind    SourceKind
	SourceID      uuid.UUID
	ResultID      uuid.UUID
	DeliveryJobID *uuid.UUID
	OccurredAt    time.Time
}

type IngestResult struct {
	RootID     uuid.UUID
	Created    bool
	RecordedAt time.Time
}
