package iam

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityUserCreateRequestHashCanonicalizesRoleOrder(t *testing.T) {
	firstRole := uuid.MustParse("81000000-0000-4000-8000-000000000001")
	secondRole := uuid.MustParse("82000000-0000-4000-8000-000000000001")
	base := IdentityUserCreate{
		Email: "user@example.com", FirstName: "Ada", LastName: "Lovelace",
		OrganizationID: uuid.MustParse("83000000-0000-4000-8000-000000000001"),
		RoleIDs:        []uuid.UUID{secondRole, firstRole},
	}
	credential := identityPreparedUserCredential{passwordSHA256: sha256.Sum256([]byte("secret-one"))}
	first, err := identityUserCreateRequestSHA(base, []byte(`{}`), credential)
	if err != nil {
		t.Fatal(err)
	}
	reordered := base
	reordered.RoleIDs = []uuid.UUID{firstRole, secondRole}
	second, err := identityUserCreateRequestSHA(reordered, []byte(`{}`), credential)
	if err != nil || second != first {
		t.Fatalf("reordered digest=%x want=%x error=%v", second, first, err)
	}
	credential.passwordSHA256 = sha256.Sum256([]byte("secret-two"))
	changed, err := identityUserCreateRequestSHA(reordered, []byte(`{}`), credential)
	if err != nil || changed == first {
		t.Fatal("password digest is not bound to the idempotency request hash")
	}
}

func TestIdentitySortedIDsDoesNotMutateCaller(t *testing.T) {
	first := uuid.MustParse("84000000-0000-4000-8000-000000000001")
	second := uuid.MustParse("85000000-0000-4000-8000-000000000001")
	input := []uuid.UUID{second, first}
	result := identitySortedIDs(input)
	if input[0] != second || result[0] != first || result[1] != second {
		t.Fatalf("input=%v result=%v", input, result)
	}
}
