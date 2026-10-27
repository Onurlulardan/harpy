package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestReadPreferencesRejectsMissingRecipient(t *testing.T) {
	if _, err := (&Store{}).ReadPreferences(t.Context(), domain.ID{}); err != ErrInvalid {
		t.Fatalf("ReadPreferences() error = %v", err)
	}
}
