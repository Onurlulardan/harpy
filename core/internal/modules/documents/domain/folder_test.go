package domain

import (
	"errors"
	"testing"
	"time"
)

func TestFolderMutationsUseRevisionAndStableIdentity(t *testing.T) {
	now := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	folder, err := NewFolder(
		mustID(t, "018f0000-0000-7000-8000-000000000101"),
		mustID(t, "11111111-1111-4111-8111-111111111111"), ID{},
		mustID(t, "22222222-2222-4222-8222-222222222222"), " Contracts ", now,
	)
	if err != nil || folder.DisplayName != "Contracts" || folder.RowVersion != 1 {
		t.Fatalf("folder = %+v, %v", folder, err)
	}
	renamed, err := folder.Rename(1, "Legal", now.Add(time.Second))
	if err != nil || renamed.RowVersion != 2 || renamed.ID != folder.ID {
		t.Fatalf("renamed = %+v, %v", renamed, err)
	}
	if _, err = renamed.Move(1, ID{}, now.Add(2*time.Second)); !errors.Is(err, ErrFolderTransition) {
		t.Fatalf("stale revision = %v", err)
	}
	trashed, err := renamed.Trash(2, now.Add(2*time.Second))
	if err != nil || trashed.State != FolderTrashed {
		t.Fatalf("trashed = %+v, %v", trashed, err)
	}
	restored, err := trashed.Restore(3, now.Add(3*time.Second))
	if err != nil || restored.State != FolderActive || restored.RowVersion != 4 {
		t.Fatalf("restored = %+v, %v", restored, err)
	}
}

func TestFolderRejectsSelfParentAndControlName(t *testing.T) {
	now := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	id := mustID(t, "018f0000-0000-7000-8000-000000000101")
	_, err := NewFolder(id, mustID(t, "11111111-1111-4111-8111-111111111111"), id,
		mustID(t, "22222222-2222-4222-8222-222222222222"), "bad\nname", now)
	if !errors.Is(err, ErrFolderInvalid) {
		t.Fatalf("self parent/control name error = %v", err)
	}
}
