package application

import "testing"

func TestRewrapProviderSecretRejectsEqualKEKRevisions(t *testing.T) {
	application, _, scope := providerFixture(t)
	err := application.RewrapProviderSecret(t.Context(), scope,
		RewrapProviderSecretInput{TargetKEKRevision: 1, ExpectedKEKRevision: 1})
	if err != ErrInvalid {
		t.Fatalf("RewrapProviderSecret() error = %v", err)
	}
}
