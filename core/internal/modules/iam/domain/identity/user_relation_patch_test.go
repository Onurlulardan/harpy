package identity

import (
	"errors"
	"testing"
)

func TestUserRelationPatchCanonicalizesAndPairsStableRows(t *testing.T) {
	userID := IdentityID{9}
	high := IdentityID{2}
	low := IdentityID{1}
	first, err := NewOrganizationSetPatch(
		userID, true, []IdentityID{high, low},
	)
	if err != nil {
		t.Fatalf("first patch: %v", err)
	}
	second, err := NewOrganizationSetPatch(
		userID, true, []IdentityID{low, high},
	)
	if err != nil {
		t.Fatalf("second patch: %v", err)
	}
	if got := first.OrganizationIDs(); len(got) != 2 || got[0] != low || got[1] != high {
		t.Fatalf("canonical organizations = %v", got)
	}
	firstRows := first.MembershipIDs()
	secondRows := second.MembershipIDs()
	wantFirstRow := IdentityID{
		0xdf, 0x2d, 0xdc, 0x54, 0x55, 0x09, 0x80, 0xa6,
		0xb5, 0x97, 0xd4, 0x6d, 0xb6, 0x20, 0x19, 0xb3,
	}
	if len(firstRows) != 2 || firstRows[0] == (IdentityID{}) ||
		firstRows[0] != secondRows[0] || firstRows[1] != secondRows[1] ||
		firstRows[0] == firstRows[1] || firstRows[0] != wantFirstRow {
		t.Fatalf("deterministic membership rows = %v / %v", firstRows, secondRows)
	}
	firstRows[0] = IdentityID{}
	if first.MembershipIDs()[0] == (IdentityID{}) {
		t.Fatal("patch exposed mutable row IDs")
	}
}

func TestUserRelationPatchAcceptsExplicitEmptySet(t *testing.T) {
	patch, err := NewRoleSetPatch(IdentityID{1}, true, nil)
	if err != nil || !patch.Replace() || len(patch.RoleIDs()) != 0 ||
		len(patch.AssignmentIDs()) != 0 {
		t.Fatalf("explicit empty patch = %+v, %v", patch, err)
	}
	if _, err = NewRoleSetPatch(IdentityID{1}, false, []IdentityID{{2}}); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("omitted patch with values error = %v", err)
	}
}

func TestUserRelationPatchRejectsDuplicateAndOversizedSets(t *testing.T) {
	userID := IdentityID{1}
	targetID := IdentityID{2}
	if _, err := NewRoleSetPatch(
		userID, true, []IdentityID{targetID, targetID},
	); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("duplicate error = %v", err)
	}
	oversized := make([]IdentityID, MaximumUserRelationSetSize+1)
	for index := range oversized {
		oversized[index][14] = byte(index/255 + 1)
		oversized[index][15] = byte(index + 1)
	}
	if _, err := NewOrganizationSetPatch(userID, true, oversized); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("oversized error = %v", err)
	}
}
