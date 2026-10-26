//go:build integration

package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func iamAuthLocatorSourceSHA256(
	t *testing.T,
	tokenID string,
	familyID string,
	selector []byte,
	revision int64,
	state string,
	secretHMAC []byte,
) []byte {
	t.Helper()
	if len(selector) != 16 || len(secretHMAC) != 32 {
		t.Fatal("IAM auth locator digest input length drifted")
	}
	var selectorValue [16]byte
	var secretValue [32]byte
	copy(selectorValue[:], selector)
	copy(secretValue[:], secretHMAC)
	digest, err := authentication.RefreshLocatorSourceSHA256(
		uuid.MustParse(tokenID), uuid.MustParse(familyID), selectorValue,
		revision, state, secretValue,
	)
	if err != nil {
		t.Fatalf("IAM auth locator digest: %v", err)
	}
	return digest[:]
}
