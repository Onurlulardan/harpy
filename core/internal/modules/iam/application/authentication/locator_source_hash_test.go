package authentication

import (
	"testing"

	"github.com/google/uuid"
)

func TestRefreshLocatorSourceSHA256BindsIdentityRevisionAndState(t *testing.T) {
	tokenID := uuid.MustParse("64000000-0000-4000-8000-000000000001")
	familyID := uuid.MustParse("63000000-0000-4000-8000-000000000001")
	selector := [16]byte{1}
	secret := [32]byte{2}
	active, err := RefreshLocatorSourceSHA256(
		tokenID, familyID, selector, 1, "active", secret,
	)
	if err != nil || active == [32]byte{} {
		t.Fatalf("active locator digest: %x %v", active, err)
	}
	consumed, err := RefreshLocatorSourceSHA256(
		tokenID, familyID, selector, 2, "consumed", secret,
	)
	if err != nil || consumed == active {
		t.Fatalf("consumed locator digest did not advance: %x %v", consumed, err)
	}
	if _, err := RefreshLocatorSourceSHA256(
		uuid.Nil, familyID, selector, 1, "active", secret,
	); err != ErrAuthenticationContract {
		t.Fatalf("invalid locator digest error = %v", err)
	}
}
