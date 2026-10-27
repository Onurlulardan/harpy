package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestGetMessageRejectsMissingIdentity(t *testing.T) {
	if _, err := (&Store{}).GetMessage(t.Context(), domain.ID{}, domain.ID{}); err != ErrInvalid {
		t.Fatalf("GetMessage() error = %v", err)
	}
}
