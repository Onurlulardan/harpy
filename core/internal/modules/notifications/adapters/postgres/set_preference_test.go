package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestSetPreferenceRejectsInvalidPreference(t *testing.T) {
	if err := (&Store{}).SetPreference(t.Context(), domain.Preference{}, 0); err != ErrInvalid {
		t.Fatalf("SetPreference() error = %v", err)
	}
}
