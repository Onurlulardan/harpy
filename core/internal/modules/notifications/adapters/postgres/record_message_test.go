package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestRecordMessageRejectsInvalidMessage(t *testing.T) {
	if err := (&Store{}).RecordMessage(t.Context(), domain.Message{}); err != ErrInvalid {
		t.Fatalf("RecordMessage() error = %v", err)
	}
}
