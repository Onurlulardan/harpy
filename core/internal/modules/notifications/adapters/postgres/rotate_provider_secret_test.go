package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestRotateProviderSecretRejectsInvalidSecret(t *testing.T) {
	if err := (&Store{}).RotateProviderSecret(t.Context(), ports.ProviderConfig{},
		ports.EncryptedSecret{}, 0); err != ErrInvalid {
		t.Fatalf("RotateProviderSecret() error = %v", err)
	}
}
