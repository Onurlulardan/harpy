package api

import (
	"context"
)

type Scope interface {
	OperationID() string
}

type NotificationIntentPort interface {
	RecordIntent(context.Context, Scope, IntentRequest) (IntentResult, error)
}

type IntentRequest struct {
	Producer, NotificationType string
	RecipientPrincipalID       ID
	TemplateVersionID          ID
	ProviderConfigVersionID    ID
	LocaleInput, EffectKey     string
	Category                   Category
	Channels                   []Channel
	Data                       []TemplateValue
	DataDigest                 [32]byte
	CorrelationID, CausationID ID
}

type IntentResult struct {
	MessageID      ID
	MigrationJobID ID
	InboxRevision  int64
	Replay         bool
	Outcome        string
}
