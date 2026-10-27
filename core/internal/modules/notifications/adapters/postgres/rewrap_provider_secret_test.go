package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestRewrapProviderSecretRejectsInvalidSecret(t *testing.T) {
	if err := (&Store{}).RewrapProviderSecret(t.Context(), ports.EncryptedSecret{}, 0); err != ErrInvalid {
		t.Fatalf("RewrapProviderSecret() error = %v", err)
	}
}
