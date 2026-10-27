package application

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type testSealer struct{}

func (testSealer) Seal(_ context.Context, id domain.ID, plaintext []byte,
	revision int64,
) (ports.EncryptedSecret, error) {
	return ports.EncryptedSecret{ID: id, Ciphertext: append([]byte(nil), plaintext...),
		Nonce: []byte{1}, WrappedDEK: []byte{2}, KEKRevision: revision}, nil
}

func (testSealer) Rewrap(_ context.Context, secret ports.EncryptedSecret,
	revision int64,
) (ports.EncryptedSecret, error) {
	secret.WrappedDEK = []byte{byte(revision)}
	secret.KEKRevision = revision
	return secret, nil
}

func providerFixture(t *testing.T) (*ProviderApplication, *testStores, testScope) {
	t.Helper()
	base, stores, scope, _ := newIntentFixture(t)
	application, err := NewProviderApplication(base, testSealer{})
	if err != nil {
		t.Fatal(err)
	}
	stores.provider = ports.ProviderConfig{}
	return application, stores, scope
}

func TestCreateProviderConfigSupportsUpdateRotateAndRewrapLifecycle(t *testing.T) {
	application, stores, scope := providerFixture(t)
	secret := []byte("credential-one")
	created, err := application.CreateProviderConfig(t.Context(), scope,
		CreateProviderConfigInput{Kind: ports.ProviderSMTP, Name: "Tenant SMTP",
			Enabled: true, Secret: secret, KEKRevision: 3,
			Metadata: ports.ProviderMetadata{SMTPHost: "smtp.example.test", SMTPPort: 465,
				SMTPSecurity: "implicit_tls", FromAddress: "Forge <noreply@example.test>"}})
	if err != nil || created.Revision != 1 || !created.SecretID.Valid() {
		t.Fatalf("CreateProviderConfig() = %+v, %v", created, err)
	}
	for _, value := range secret {
		if value != 0 {
			t.Fatal("caller secret was not cleared")
		}
	}
	updated, err := application.UpdateProviderConfig(t.Context(), scope,
		UpdateProviderConfigInput{ID: created.ID, Name: "Primary SMTP", Enabled: true,
			ExpectedRevision: 1, Metadata: created.Metadata})
	if err != nil || updated.Revision != 2 || updated.VersionID == created.VersionID {
		t.Fatalf("UpdateProviderConfig() = %+v, %v", updated, err)
	}
	rotated, err := application.RotateProviderSecret(t.Context(), scope,
		RotateProviderSecretInput{ID: created.ID, Secret: []byte("credential-two"),
			KEKRevision: 3, ExpectedRevision: 2})
	if err != nil || rotated.Revision != 3 || rotated.SecretID == created.SecretID {
		t.Fatalf("RotateProviderSecret() = %+v, %v", rotated, err)
	}
	if err = application.RewrapProviderSecret(t.Context(), scope,
		RewrapProviderSecretInput{ID: created.ID, TargetKEKRevision: 4,
			ExpectedKEKRevision: 3}); err != nil {
		t.Fatal(err)
	}
	rewrapped := stores.providerSecrets[rotated.SecretID]
	if rewrapped.KEKRevision != 4 || len(rewrapped.WrappedDEK) != 1 ||
		rewrapped.WrappedDEK[0] != 4 {
		t.Fatalf("rewrapped secret = %+v", rewrapped)
	}
}

func TestProviderMetadataRejectsUnsafeEndpoints(t *testing.T) {
	unsafe := []ports.ProviderMetadata{
		{SMTPHost: "127.0.0.1", SMTPPort: 25, SMTPSecurity: "starttls",
			FromAddress: "noreply@example.test"},
		{TelegramAPIBase: "https://evil.example"},
	}
	if validateProviderMetadata(ports.ProviderSMTP, unsafe[0]) ||
		validateProviderMetadata(ports.ProviderTelegram, unsafe[1]) {
		t.Fatal("unsafe provider metadata accepted")
	}
}
