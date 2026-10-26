package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestCreateProviderConfigRejectsInvalidConfiguration(t *testing.T) {
	if err := (&Store{}).CreateProviderConfig(t.Context(), ports.ProviderConfig{}); err != ErrInvalid {
		t.Fatalf("CreateProviderConfig() error = %v", err)
	}
}
