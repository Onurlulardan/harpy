package lifecycle

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
)

func TestCredentialRequestSHA256BindsPasswordWithoutPlainDigest(t *testing.T) {
	presented := credentialRequestFixture(t, 0x31, 0x41)
	defer presented.Destroy()
	tenantID := uuid.MustParse("71000000-0000-4000-8000-000000000001")
	tokenID := uuid.MustParse("72000000-0000-4000-8000-000000000001")
	password := []byte("request-password-value")

	first, err := presented.RequestSHA256(
		tenantID, PurposePasswordReset, tokenID, password,
	)
	if err != nil {
		t.Fatal(err)
	}
	second, err := presented.RequestSHA256(
		tenantID, PurposePasswordReset, tokenID, password,
	)
	if err != nil || first != second {
		t.Fatalf("stable request binding mismatch: %v", err)
	}
	plain := sha256.Sum256(password)
	if first == plain || bytes.Contains(first[:], password) {
		t.Fatal("request binding exposed a reusable plaintext password digest")
	}
}

func TestCredentialRequestSHA256SeparatesDifferentPasswords(t *testing.T) {
	presented := credentialRequestFixture(t, 0x32, 0x42)
	tenantID := uuid.MustParse("71000000-0000-4000-8000-000000000002")
	tokenID := uuid.MustParse("72000000-0000-4000-8000-000000000002")
	first, err := presented.RequestSHA256(
		tenantID, PurposePasswordReset, tokenID, []byte("first-password"),
	)
	if err != nil {
		t.Fatal(err)
	}
	second, err := presented.RequestSHA256(
		tenantID, PurposePasswordReset, tokenID, []byte("second-password"),
	)
	if err != nil || first == second {
		t.Fatalf("different password reused request identity: %v", err)
	}
	presented.Destroy()
	if _, err := presented.RequestSHA256(
		tenantID, PurposePasswordReset, tokenID, []byte("first-password"),
	); err == nil {
		t.Fatal("destroyed credential produced a request identity")
	}
}

func credentialRequestFixture(
	t *testing.T,
	selectorByte byte,
	secretByte byte,
) *PresentedCredential {
	t.Helper()
	selector := bytes.Repeat([]byte{selectorByte}, 16)
	secret := bytes.Repeat([]byte{secretByte}, 32)
	wire := []byte(base64.RawURLEncoding.EncodeToString(selector) + "." +
		base64.RawURLEncoding.EncodeToString(secret))
	presented, err := ParseCredential(wire)
	clear(wire)
	clear(selector)
	clear(secret)
	if err != nil {
		t.Fatal(err)
	}
	return presented
}
