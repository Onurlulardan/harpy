package cryptography

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

type totpDecryptorStub struct {
	secret [32]byte
	err    error
}

func (stub *totpDecryptorStub) DecryptTOTP(
	_ context.Context, _ string, _, _, _ []byte, _ [32]byte,
) ([32]byte, error) {
	return stub.secret, stub.err
}

func TestTOTPVerifiesRFC6238SHA256SixDigitWindow(t *testing.T) {
	decryptor := &totpDecryptorStub{}
	copy(decryptor.secret[:], []byte("12345678901234567890123456789012"))
	provider, err := NewTOTP(decryptor)
	if err != nil {
		t.Fatal(err)
	}
	envelope := validTOTPEnvelope()
	code, err := NewTOTPCode([]byte("119246"))
	if err != nil {
		t.Fatal(err)
	}
	defer code.Destroy()
	step, err := provider.Verify(
		context.Background(), code, envelope, time.Unix(59, 0).UTC(),
	)
	if err != nil || step != 1 {
		t.Fatalf("RFC6238 step=%d err=%v", step, err)
	}
	previous := totpDigits(decryptor.secret, 1)
	previousCode, _ := NewTOTPCode(previous[:])
	defer previousCode.Destroy()
	step, err = provider.Verify(
		context.Background(), previousCode, envelope, time.Unix(60, 0).UTC(),
	)
	if err != nil || step != 1 {
		t.Fatalf("previous TOTP window step=%d err=%v", step, err)
	}
}

func TestTOTPRejectsInvalidCodeAndDecryptFailure(t *testing.T) {
	decryptor := &totpDecryptorStub{secret: [32]byte{1}}
	provider, _ := NewTOTP(decryptor)
	code, _ := NewTOTPCode([]byte("000000"))
	defer code.Destroy()
	_, err := provider.Verify(
		context.Background(), code, validTOTPEnvelope(), time.Unix(90, 0),
	)
	if !errors.Is(err, authentication.ErrInvalidCredentials) {
		t.Fatalf("invalid TOTP error=%v", err)
	}
	decryptor.err = errors.New("kms unavailable")
	_, err = provider.Verify(
		context.Background(), code, validTOTPEnvelope(), time.Unix(90, 0),
	)
	if !errors.Is(err, authentication.ErrAuthenticationCrypto) {
		t.Fatalf("decrypt failure error=%v", err)
	}
}

func validTOTPEnvelope() authentication.TOTPEnvelope {
	return authentication.TOTPEnvelope{
		Profile: "totp_sha256_6_30_v1", CipherSuite: "aes_256_gcm_v1",
		KMSKeyID: "kms-v1", Nonce: bytes.Repeat([]byte{1}, 12),
		Ciphertext:        bytes.Repeat([]byte{2}, 32),
		AuthenticationTag: bytes.Repeat([]byte{3}, 16),
		AADSHA256:         [32]byte{4}, Revision: 1,
	}
}
