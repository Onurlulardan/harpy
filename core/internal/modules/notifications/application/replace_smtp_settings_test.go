package application

import (
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func smtpSettingsFixture(t *testing.T) (
	*SMTPSettingsApplication, *testStores, testScope,
) {
	t.Helper()
	base, stores, scope, _ := newIntentFixture(t)
	stores.provider = ports.ProviderConfig{}
	application, err := NewSMTPSettingsApplication(base, testSealer{}, 3)
	if err != nil {
		t.Fatal(err)
	}
	return application, stores, scope
}

func smtpReplaceRequest(password string) api.SMTPSettingsReplaceRequest {
	return api.SMTPSettingsReplaceRequest{Host: "smtp.example.test", Port: 587,
		Username: "tenant-mailer", Password: []byte(password),
		From: "Forge <noreply@example.test>", ExpectedSettingsRevision: 7,
		ResultSettingsRevision: 8, IdempotencyKey: "tenant-settings-0001",
		RequestSHA256: sha256.Sum256([]byte("tenant settings request"))}
}

func TestReplaceSMTPSettingsCreatesRotatesAndReplaysExactly(t *testing.T) {
	application, stores, scope := smtpSettingsFixture(t)
	first := smtpReplaceRequest("smtp-password-one")
	result, err := application.ReplaceSMTPSettings(t.Context(), scope, first)
	if err != nil || result.Replay || result.ResultSettingsRevision != 8 {
		t.Fatalf("first replace = %+v, %v", result, err)
	}
	for _, value := range first.Password {
		if value != 0 {
			t.Fatal("caller password was not cleared")
		}
	}
	created := stores.provider
	if created.Revision != 1 || created.Metadata.SMTPUsername != "tenant-mailer" ||
		string(stores.providerSecrets[created.SecretID].Ciphertext) != "smtp-password-one" {
		t.Fatalf("created provider = %+v", created)
	}
	replay, err := application.ReplaceSMTPSettings(t.Context(), scope,
		smtpReplaceRequest("smtp-password-one"))
	if err != nil || !replay.Replay || stores.provider.Revision != 1 {
		t.Fatalf("replay = %+v, %v provider=%+v", replay, err, stores.provider)
	}
	conflict := smtpReplaceRequest("smtp-password-two")
	if _, err = application.ReplaceSMTPSettings(t.Context(), scope,
		conflict); !errors.Is(err, api.ErrSMTPSettingsConflict) {
		t.Fatalf("different digest error = %v", err)
	}
}

func TestReplaceSMTPSettingsUpdatesMetadataAndRotatesCredentialAtomically(t *testing.T) {
	application, stores, scope := smtpSettingsFixture(t)
	if _, err := application.ReplaceSMTPSettings(t.Context(), scope,
		smtpReplaceRequest("smtp-password-one")); err != nil {
		t.Fatal(err)
	}
	previousSecret := stores.provider.SecretID
	next := smtpReplaceRequest("smtp-password-two")
	next.ExpectedSettingsRevision, next.ResultSettingsRevision = 8, 9
	next.IdempotencyKey = "tenant-settings-0002"
	next.Host = "mail.example.test"
	result, err := application.ReplaceSMTPSettings(t.Context(), scope, next)
	if err != nil || result.ResultSettingsRevision != 9 ||
		stores.provider.Revision != 2 || stores.provider.SecretID == previousSecret ||
		stores.provider.Metadata.SMTPHost != "mail.example.test" {
		t.Fatalf("rotated result=%+v provider=%+v err=%v", result, stores.provider, err)
	}
}
