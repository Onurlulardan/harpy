package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type SMTPSettingsReceipt struct {
	IdempotencyKey                                   string
	RequestDigest                                    [32]byte
	ExpectedSettingsRevision, ResultSettingsRevision int64
	ProviderConfigID                                 domain.ID
	ProviderRevision                                 int64
}

type SMTPTestIntent struct {
	OperationID, ProviderVersionID, ActorID domain.ID
	IdempotencyKey                          string
	RequestDigest                           [32]byte
	ExpectedSettingsRevision                int64
	EncryptedTarget                         EncryptedSecret
	RequestedAt                             time.Time
}

type SMTPSettingsStore interface {
	LockSMTPSettings(context.Context) error
	LatestSMTPSettingsRevision(context.Context) (int64, error)
	FindSMTPSettingsReceipt(context.Context, string) (SMTPSettingsReceipt, bool, error)
	InsertSMTPSettingsReceipt(context.Context, SMTPSettingsReceipt) error
	FindSMTPTestIntent(context.Context, string) (SMTPTestIntent, bool, error)
	InsertSMTPTestIntent(context.Context, SMTPTestIntent) error
}
