package v1adoption

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Run struct {
	ID         uuid.UUID
	State      string
	Checkpoint Checkpoint
	Unresolved int64
}

type Checkpoint struct {
	LastSourceID   *uuid.UUID
	AdoptedCount   int64
	SourceChecksum [32]byte
	TargetChecksum [32]byte
}

type SourceRecord struct {
	ID, TenantID       uuid.UUID
	UserID             *uuid.UUID
	NotificationType   string
	Level, Title, Body string
	Action, Entity     string
	EntityID           *uuid.UUID
	Data               json.RawMessage
	Channel            string
	Read               bool
	CreatedAt          *time.Time
	ReadAt             *time.Time
	InboxRevision      int64
}

type SourcePage struct {
	Records []SourceRecord
	Final   bool
}

type MappedRecord struct {
	MessageID, RecipientID, DeliveryID, AttemptID uuid.UUID
	NotificationType, Category, Subject, Body     string
	Channel                                       string
	InboxRevision                                 int64
	CreatedAt                                     time.Time
	ReadAt                                        *time.Time
	CanonicalDigest, EffectKeyHash                [32]byte
	SubjectDigest, BodyDigest                     [32]byte
	ProviderReferenceDigest                       [32]byte
}

type StepResult struct {
	RunID        uuid.UUID
	AdoptedCount int64
	Processed    int
	Complete     bool
}
