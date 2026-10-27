package domain

import (
	"bytes"
	"testing"
	"time"
)

func messageFixture(t *testing.T) Message {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	id, _ := NewV7(now, bytes.NewReader(bytes.Repeat([]byte{1}, 10)))
	recipient, _ := ParseID("11111111-1111-4111-8111-111111111111")
	template, _ := ParseID("22222222-2222-4222-8222-222222222222")
	value, err := NewMessage(Message{ID: id, RecipientID: recipient,
		TemplateVersionID: template, Producer: "workflow", NotificationType: "task.due",
		Category: CategoryWorkflowAction, TemplateKind: TemplatePlainText,
		Locale: "tr", Subject: "Konu", Body: "Gövde",
		Revision: 1, InboxRevision: 7, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestMessageAcceptsStableLegacyIdentity(t *testing.T) {
	messageID, _ := ParseID("10000000-0000-4000-8000-000000000001")
	recipient, _ := ParseID("10000000-0000-4000-8000-000000000002")
	template, _ := ParseID("10000000-0000-4000-8000-000000000003")
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	message, err := NewMessage(Message{ID: messageID, RecipientID: recipient,
		TemplateVersionID: template, Producer: "legacy_v1", NotificationType: "legacy.notice",
		Category: CategorySystem, TemplateKind: TemplatePlainText, Locale: "tr",
		Subject: "subject", Body: "body", Revision: 1, InboxRevision: 1, CreatedAt: now})
	if err != nil || message.ID != messageID {
		t.Fatalf("stable legacy identity rejected: %#v %v", message, err)
	}
}

func TestMessageReadDismissAndPurgeAreRevisionSafe(t *testing.T) {
	now := time.Date(2026, 7, 18, 13, 0, 0, 0, time.UTC)
	value := messageFixture(t)
	read, err := value.MarkRead(1, now)
	if err != nil || read.Revision != 2 || read.ReadAt == nil {
		t.Fatalf("MarkRead() = %+v, %v", read, err)
	}
	if _, err = read.Dismiss(1, now); !errorsIs(err, ErrMessageConflict) {
		t.Fatalf("stale dismiss error = %v", err)
	}
	dismissed, err := read.Dismiss(2, now)
	if err != nil || dismissed.DismissedAt == nil {
		t.Fatalf("Dismiss() = %+v, %v", dismissed, err)
	}
	purged, err := dismissed.PurgeContent(now)
	if err != nil || purged.Body != "" || purged.PurgedAt == nil {
		t.Fatalf("PurgeContent() = %+v, %v", purged, err)
	}
}

func errorsIs(got, want error) bool { return got == want }
