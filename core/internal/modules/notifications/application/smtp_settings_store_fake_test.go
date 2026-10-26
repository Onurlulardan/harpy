package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (*testStores) LockSMTPSettings(context.Context) error { return nil }

func (value *testStores) LatestSMTPSettingsRevision(context.Context) (int64, error) {
	return value.latestSMTPRevision, nil
}

func (value *testStores) FindSMTPSettingsReceipt(_ context.Context,
	key string,
) (ports.SMTPSettingsReceipt, bool, error) {
	receipt, found := value.smtpReceipts[key]
	return receipt, found, nil
}

func (value *testStores) InsertSMTPSettingsReceipt(_ context.Context,
	receipt ports.SMTPSettingsReceipt,
) error {
	value.smtpReceipts[receipt.IdempotencyKey] = receipt
	value.latestSMTPRevision = receipt.ResultSettingsRevision
	return nil
}

func (value *testStores) FindSMTPTestIntent(_ context.Context,
	key string,
) (ports.SMTPTestIntent, bool, error) {
	intent, found := value.smtpTestIntents[key]
	return intent, found, nil
}

func (value *testStores) InsertSMTPTestIntent(_ context.Context,
	intent ports.SMTPTestIntent,
) error {
	value.smtpTestIntents[intent.IdempotencyKey] = intent
	return nil
}
