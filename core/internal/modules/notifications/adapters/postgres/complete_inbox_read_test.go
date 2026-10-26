package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestCompleteInboxReadRejectsInvalidBoundary(t *testing.T) {
	if _, err := (&Store{}).CompleteInboxRead(t.Context(), domain.ID{}, 0,
		domain.ID{}); err != ErrInvalid {
		t.Fatalf("CompleteInboxRead() error = %v", err)
	}
}
