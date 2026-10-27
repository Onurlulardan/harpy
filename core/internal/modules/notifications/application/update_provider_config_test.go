package application

import "testing"

func TestUpdateProviderConfigRejectsMissingIdentity(t *testing.T) {
	application, _, scope := providerFixture(t)
	if _, err := application.UpdateProviderConfig(t.Context(), scope,
		UpdateProviderConfigInput{Name: "SMTP", ExpectedRevision: 1}); err != ErrInvalid {
		t.Fatalf("UpdateProviderConfig() error = %v", err)
	}
}
