package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestUpdateProviderConfigRejectsInvalidRevision(t *testing.T) {
	if err := (&Store{}).UpdateProviderConfig(t.Context(), ports.ProviderConfig{}, 0); err != ErrInvalid {
		t.Fatalf("UpdateProviderConfig() error = %v", err)
	}
}
