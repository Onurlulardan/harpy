package application

import "testing"

func TestRotateProviderSecretRejectsEmptySecret(t *testing.T) {
	application, _, scope := providerFixture(t)
	if _, err := application.RotateProviderSecret(t.Context(), scope,
		RotateProviderSecretInput{ExpectedRevision: 1, KEKRevision: 1}); err != ErrInvalid {
		t.Fatalf("RotateProviderSecret() error = %v", err)
	}
}
