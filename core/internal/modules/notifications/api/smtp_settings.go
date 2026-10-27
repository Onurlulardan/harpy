package api

import (
	"context"
	"time"
)

type SMTPSettingsPort interface {
	ReadSMTPSettings(context.Context, Scope) (SMTPSettingsProjection, error)
	ReplaceSMTPSettings(context.Context, Scope,
		SMTPSettingsReplaceRequest) (SMTPSettingsReplaceResult, error)
	EnqueueSMTPTest(context.Context, Scope,
		SMTPTestIntentRequest) (SMTPTestIntentResult, error)
}

type SMTPSettingsProjection struct {
	Configured         bool
	Host               string
	Port               int
	Username           string
	From               string
	PasswordConfigured bool
}

type SMTPSettingsReplaceRequest struct {
	Host, Username, From     string
	Port                     int
	Password                 []byte
	ExpectedSettingsRevision int64
	ResultSettingsRevision   int64
	IdempotencyKey           string
	RequestSHA256            [32]byte
}

type SMTPSettingsReplaceResult struct {
	ResultSettingsRevision int64
	Replay                 bool
}

type SMTPTestIntentRequest struct {
	To                       string
	ExpectedSettingsRevision int64
	IdempotencyKey           string
	RequestSHA256            [32]byte
	RequestedAt              time.Time
}

type SMTPTestIntentResult struct {
	OperationID ID
	Replay      bool
}
