package application

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestGetMessageScopesToVerifiedActor(t *testing.T) {
	application, _, scope, request := newIntentFixture(t)
	created, err := application.RecordIntent(t.Context(), scope, request)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := domainID(created.MessageID)
	if _, err = application.GetMessage(t.Context(), scope, id); err != nil {
		t.Fatal(err)
	}
	scope.actor = uuid.UUID(request.CorrelationID)
	if _, err = application.GetMessage(t.Context(), scope, id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-recipient read error = %v", err)
	}
}
