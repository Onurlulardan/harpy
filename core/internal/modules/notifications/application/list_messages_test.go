package application

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestListMessagesRejectsUnboundedPage(t *testing.T) {
	application, _, scope, _ := newIntentFixture(t)
	if _, err := application.ListMessages(t.Context(), scope,
		ports.MessageFilter{Limit: 101}); err != ErrInvalid {
		t.Fatalf("unbounded list error = %v", err)
	}
}
