package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestResolveRecipientRejectsInvalidBinder(t *testing.T) {
	var binder *Binder
	if _, err := binder.Recipient(t.Context(), nil, domain.ID{}); err != ErrInvalid {
		t.Fatalf("Recipient() error = %v", err)
	}
}
