package password

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestAdoptedBcryptVerifiesOnlyBeforeRetirementDeadline(t *testing.T) {
	adoptedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	hash, err := bcrypt.GenerateFromPassword([]byte("v1-only-secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := LegacyBcryptCredentialFromStorage(string(hash), adoptedAt)
	if err != nil {
		t.Fatal(err)
	}
	if credential.Algorithm() != AlgorithmBcryptV1Adopted ||
		!credential.Valid() || !credential.NeedsRehash() {
		t.Fatal("adopted bcrypt credential metadata is invalid")
	}
	secret, err := NewVerificationSecret([]byte("v1-only-secret"))
	if err != nil {
		t.Fatal(err)
	}
	defer secret.Destroy()
	matched, err := VerifyLegacy(secret, credential, adoptedAt.AddDate(0, 0, 179))
	if err != nil || !matched {
		t.Fatalf("legacy credential did not verify: matched=%v err=%v", matched, err)
	}
	matched, err = VerifyLegacy(secret, credential, adoptedAt.AddDate(0, 0, 180))
	if matched || !errors.Is(err, ErrLegacyCredentialExpired) {
		t.Fatalf("retired legacy credential result: matched=%v err=%v", matched, err)
	}
}

func TestAdoptedBcryptRejectsMalformedAndWrongCredentials(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if _, err := LegacyBcryptCredentialFromStorage("$2b$broken", now); err == nil {
		t.Fatal("malformed bcrypt credential was accepted")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("expected"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := LegacyBcryptCredentialFromStorage(string(hash), now)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := NewVerificationSecret([]byte("wrong"))
	if err != nil {
		t.Fatal(err)
	}
	defer secret.Destroy()
	matched, err := VerifyLegacy(secret, credential, now)
	if err != nil || matched {
		t.Fatalf("wrong credential result: matched=%v err=%v", matched, err)
	}
}
