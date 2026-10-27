package application

import "testing"

func TestCompleteInboxReadUsesMonotonicInboxBoundary(t *testing.T) {
	application, stores, scope, request := newIntentFixture(t)
	if count, err := application.CompleteInboxRead(t.Context(), scope); err != nil || count != 0 {
		t.Fatalf("empty CompleteInboxRead() = %d, %v", count, err)
	}
	if _, err := application.RecordIntent(t.Context(), scope, request); err != nil {
		t.Fatal(err)
	}
	if count, err := application.CompleteInboxRead(t.Context(), scope); err != nil || count != 1 {
		t.Fatalf("CompleteInboxRead() = %d, %v; messages=%d", count, err, len(stores.messages))
	}
}
