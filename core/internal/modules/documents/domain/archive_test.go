package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestArchivePinsImmutableBoundedSnapshot(t *testing.T) {
	archive := validArchive(t)
	entry := ArchiveEntry{
		DocumentID:     mustID(t, "018f0000-0000-7000-8000-000000000031"),
		VersionID:      mustID(t, "018f0000-0000-7000-8000-000000000032"),
		PlaintextBytes: 7, PlaintextSHA256: sha256.Sum256([]byte("content")),
	}
	pinned, err := archive.Pin([]ArchiveEntry{entry}, sha256.Sum256([]byte("entries")))
	if err != nil || pinned.EntryCount != 1 || pinned.PlaintextBytes != 7 {
		t.Fatalf("pinned = %+v, %v", pinned, err)
	}
	tooLarge := entry
	tooLarge.PlaintextBytes = MaximumArchivePlaintext + 1
	if _, err = archive.Pin([]ArchiveEntry{tooLarge}, sha256.Sum256([]byte("large"))); !errors.Is(err, ErrArchiveLimit) {
		t.Fatalf("limit error = %v", err)
	}
}

func TestArchiveExpiresBeforeVerifiedDelete(t *testing.T) {
	archive := validArchive(t)
	entry := ArchiveEntry{
		DocumentID:     mustID(t, "018f0000-0000-7000-8000-000000000031"),
		VersionID:      mustID(t, "018f0000-0000-7000-8000-000000000032"),
		PlaintextBytes: 7, PlaintextSHA256: sha256.Sum256([]byte("content")),
	}
	archive, _ = archive.Pin([]ArchiveEntry{entry}, sha256.Sum256([]byte("entries")))
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	ready, err := archive.MarkReady(now, 128, sha256.Sum256([]byte("artifact")))
	if err != nil || ready.ExpiresAt != now.Add(ArchiveKeep) {
		t.Fatalf("ready = %+v, %v", ready, err)
	}
	if _, err = ready.Expire(now.Add(ArchiveKeep - time.Second)); !errors.Is(err, ErrArchiveTransition) {
		t.Fatalf("early expiry error = %v", err)
	}
	expired, _ := ready.Expire(ready.ExpiresAt)
	purged, err := expired.ConfirmDelete(
		ready.ExpiresAt, sha256.Sum256([]byte("exact object version delete")),
	)
	if err != nil || purged.State != ArchivePurged {
		t.Fatalf("purged = %+v, %v", purged, err)
	}
}

func validArchive(t *testing.T) Archive {
	t.Helper()
	archive, err := NewArchive(
		mustID(t, "018f0000-0000-7000-8000-000000000021"),
		mustID(t, "11111111-1111-4111-8111-111111111111"),
		mustID(t, "22222222-2222-4222-8222-222222222222"),
		mustID(t, "018f0000-0000-7000-8000-000000000022"), ID{},
		sha256.Sum256([]byte("archive-request")), 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}
