package domain

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestEntryRejectsSecretDisplayAndAcceptsDigestOnly(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	entry := Entry{ID: mustID(t, "019c0000-0000-7000-8000-000000000001"),
		OwnerAuditID: mustID(t, "019c0000-0000-7000-8000-000000000002"),
		ActorID:      mustID(t, "019c0000-0000-7000-8000-000000000003"),
		OwnerModule:  "forms", Operation: "forms.record.update",
		EntityType: "form_record", EntityID: "record-1",
		Action: ActionUpdate, AuthorizationVersion: 4, EventSHA256: sha256.Sum256([]byte("event")),
		OccurredAt: now, RecordedAt: now,
		FieldChanges: []FieldChange{{Field: "password", Classification: ClassificationSecret,
			BeforeSHA256: sha256.Sum256([]byte("old")), AfterSHA256: sha256.Sum256([]byte("new"))}}}
	if err := entry.Validate(); err != nil {
		t.Fatalf("digest-only secret entry rejected: %v", err)
	}
	secret := "must-not-be-visible"
	entry.FieldChanges[0].After = &secret
	if err := entry.Validate(); err == nil {
		t.Fatal("secret display value was accepted")
	}
}

func mustID(t *testing.T, value string) ID {
	t.Helper()
	id, err := ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestFilterIsBoundedAndRejectsInvalidRange(t *testing.T) {
	start := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	end := start.Add(-time.Hour)
	if err := (Filter{Page: 1, Limit: 100, Start: &start, End: &end}).Validate(); err == nil {
		t.Fatal("reversed audit range was accepted")
	}
	if err := (Filter{Page: 1001, Limit: 1}).Validate(); err == nil {
		t.Fatal("unbounded audit page was accepted")
	}
}
