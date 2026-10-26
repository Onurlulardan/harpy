package domain

import (
	"time"
)

type Action string

const (
	ActionCreate     Action = "create"
	ActionUpdate     Action = "update"
	ActionDelete     Action = "delete"
	ActionTransition Action = "transition"
	ActionPublish    Action = "publish"
	ActionExecute    Action = "execute"
	ActionGrant      Action = "grant"
	ActionRevoke     Action = "revoke"
	ActionRestore    Action = "restore"
	ActionPurge      Action = "purge"
)

type Classification string

const (
	ClassificationPublic       Classification = "public"
	ClassificationInternal     Classification = "internal"
	ClassificationConfidential Classification = "confidential"
	ClassificationSecret       Classification = "secret"
)

type FieldChange struct {
	Field                     string
	Classification            Classification
	Before, After             *string
	BeforeSHA256, AfterSHA256 [32]byte
}

type Entry struct {
	ID, OwnerAuditID, ActorID ID
	OrganizationID            *ID
	CorrelationID             *ID
	OwnerModule, Operation    string
	EntityType                string
	EntityID, FormSlug        string
	Action                    Action
	AuthorizationVersion      int64
	ReasonCode                string
	FieldChanges              []FieldChange
	EventSHA256               [32]byte
	OccurredAt, RecordedAt    time.Time
}
