package application

import (
	"strings"
	"testing"
)

func TestReadSMTPSettingsReturnsOnlyRedactedProjection(t *testing.T) {
	application, stores, scope := smtpSettingsFixture(t)
	request := smtpReplaceRequest("do-not-return-this-password")
	if _, err := application.ReplaceSMTPSettings(t.Context(), scope, request); err != nil {
		t.Fatal(err)
	}
	projection, err := application.ReadSMTPSettings(t.Context(), scope)
	if err != nil || !projection.Configured || !projection.PasswordConfigured ||
		projection.Host != "smtp.example.test" || projection.Port != 587 ||
		projection.Username != "tenant-mailer" ||
		projection.From != "Forge <noreply@example.test>" {
		t.Fatalf("projection = %+v, %v", projection, err)
	}
	if strings.Contains(strings.Join([]string{projection.Host, projection.Username,
		projection.From}, "\x00"), "do-not-return-this-password") {
		t.Fatal("SMTP password leaked into projection")
	}
	if len(stores.providerSecrets) != 1 {
		t.Fatalf("provider secret count = %d", len(stores.providerSecrets))
	}
}

func TestReplaceSMTPSettingsPreservesSealedPasswordWhenPasswordIsEmpty(t *testing.T) {
	application, stores, scope := smtpSettingsFixture(t)
	if _, err := application.ReplaceSMTPSettings(t.Context(), scope,
		smtpReplaceRequest("smtp-password")); err != nil {
		t.Fatal(err)
	}
	secretID := stores.provider.SecretID
	next := smtpReplaceRequest("")
	next.ExpectedSettingsRevision, next.ResultSettingsRevision = 8, 9
	next.IdempotencyKey = "tenant-settings-0002"
	next.From = "updated@example.test"
	if _, err := application.ReplaceSMTPSettings(t.Context(), scope, next); err != nil {
		t.Fatal(err)
	}
	if stores.provider.SecretID != secretID || len(stores.providerSecrets) != 1 ||
		stores.provider.Metadata.FromAddress != "updated@example.test" {
		t.Fatalf("password preservation provider = %+v", stores.provider)
	}
}
