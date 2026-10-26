package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestListMessagesRejectsUnscopedQuery(t *testing.T) {
	if _, err := (&Store{}).ListMessages(t.Context(), domain.ID{},
		ports.MessageFilter{Limit: 20}); err != ErrInvalid {
		t.Fatalf("ListMessages() error = %v", err)
	}
}
