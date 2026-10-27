package application

import "testing"

func TestDismissMessageWritesTombstoneInsteadOfDeleting(t *testing.T) {
	application, stores, scope, request := newIntentFixture(t)
	created, _ := application.RecordIntent(t.Context(), scope, request)
	id, _ := domainID(created.MessageID)
	dismissed, err := application.DismissMessage(t.Context(), scope, id)
	if err != nil || dismissed.DismissedAt == nil || len(stores.messages) != 1 {
		t.Fatalf("DismissMessage() = %+v, %v; rows=%d", dismissed, err, len(stores.messages))
	}
}
