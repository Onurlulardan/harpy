package application

import "testing"

func TestMarkReadIsRecipientScopedAndIdempotent(t *testing.T) {
	application, _, scope, request := newIntentFixture(t)
	created, _ := application.RecordIntent(t.Context(), scope, request)
	id, _ := domainID(created.MessageID)
	first, err := application.MarkRead(t.Context(), scope, id)
	if err != nil || first.ReadAt == nil || first.Revision != 2 {
		t.Fatalf("MarkRead() = %+v, %v", first, err)
	}
	second, err := application.MarkRead(t.Context(), scope, id)
	if err != nil || second.Revision != 2 {
		t.Fatalf("replay MarkRead() = %+v, %v", second, err)
	}
}
