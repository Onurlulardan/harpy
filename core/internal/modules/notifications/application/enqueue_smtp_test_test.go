package application

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func smtpTestRequest(to string) api.SMTPTestIntentRequest {
	return api.SMTPTestIntentRequest{To: to, ExpectedSettingsRevision: 8,
		IdempotencyKey: "tenant-test-email-0001",
		RequestSHA256:  sha256.Sum256([]byte("tenant test email request")),
		RequestedAt:    time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
}

func TestEnqueueSMTPTestPersistsEncryptedIntentAndStableUUIDv7(t *testing.T) {
	application, stores, scope := smtpSettingsFixture(t)
	if _, err := application.ReplaceSMTPSettings(t.Context(), scope,
		smtpReplaceRequest("smtp-password")); err != nil {
		t.Fatal(err)
	}
	first, err := application.EnqueueSMTPTest(t.Context(), scope,
		smtpTestRequest("admin@example.test"))
	if err != nil || first.Replay || !first.OperationID.Valid() ||
		first.OperationID[6]>>4 != 7 {
		t.Fatalf("first test intent = %+v, %v", first, err)
	}
	stored := stores.smtpTestIntents["tenant-test-email-0001"]
	if stored.OperationID.String() != first.OperationID.String() ||
		string(stored.EncryptedTarget.Ciphertext) != "admin@example.test" ||
		stored.ProviderVersionID != stores.provider.VersionID {
		t.Fatalf("stored test intent = %+v", stored)
	}
	replay, err := application.EnqueueSMTPTest(t.Context(), scope,
		smtpTestRequest("admin@example.test"))
	if err != nil || !replay.Replay || replay.OperationID != first.OperationID ||
		len(stores.smtpTestIntents) != 1 {
		t.Fatalf("replay = %+v, %v", replay, err)
	}
	if _, err = application.EnqueueSMTPTest(t.Context(), scope,
		smtpTestRequest("other@example.test")); !errors.Is(err, api.ErrSMTPSettingsConflict) {
		t.Fatalf("different digest error = %v", err)
	}
}
