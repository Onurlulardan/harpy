package fcartv1

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestEnvelopeKeyringBindsDEKToTenantArtifactAttemptAndKEK(t *testing.T) {
	key, err := NewKEK("export-artifact.v1", bytes.Repeat([]byte{0x11}, DEKBytes))
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := NewEnvelopeKeyring(key, nil,
		bytes.NewReader(bytes.Repeat([]byte{0x42}, DEKBytes+NonceBytes)))
	if err != nil {
		t.Fatal(err)
	}
	identity := envelopeIdentity()
	dek, envelope, err := keyring.Generate(context.Background(), identity)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := keyring.Unwrap(context.Background(), identity, envelope)
	if err != nil || opened != dek {
		t.Fatalf("unwrap mismatch: %v", err)
	}
	identity.AttemptGeneration++
	if _, err := keyring.Unwrap(context.Background(), identity, envelope); err != ErrKeyAuthentication {
		t.Fatalf("cross-attempt unwrap error = %v", err)
	}
	envelope.Ciphertext[0] ^= 1
	identity.AttemptGeneration--
	if _, err := keyring.Unwrap(context.Background(), identity, envelope); err != ErrKeyAuthentication {
		t.Fatalf("tampered unwrap error = %v", err)
	}
}

func TestEnvelopeKeyringRotationReadsPreviousKEK(t *testing.T) {
	oldKey, _ := NewKEK("export-artifact.v1", bytes.Repeat([]byte{0x11}, DEKBytes))
	newKey, _ := NewKEK("export-artifact.v2", bytes.Repeat([]byte{0x22}, DEKBytes))
	oldRing, _ := NewEnvelopeKeyring(oldKey, nil,
		bytes.NewReader(bytes.Repeat([]byte{0x33}, DEKBytes+NonceBytes)))
	identity := envelopeIdentity()
	dek, envelope, err := oldRing.Generate(context.Background(), identity)
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := NewEnvelopeKeyring(newKey, []KEK{oldKey}, nil)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := rotated.Unwrap(context.Background(), identity, envelope)
	if err != nil || opened != dek {
		t.Fatalf("rotated unwrap mismatch: %v", err)
	}
	withoutOld, _ := NewEnvelopeKeyring(newKey, nil, nil)
	if _, err := withoutOld.Unwrap(context.Background(), identity, envelope); err != ErrKEKUnavailable {
		t.Fatalf("retired key error = %v", err)
	}
}

func TestEnvelopeSecretsRedactFormattingAndOwnCiphertext(t *testing.T) {
	material := bytes.Repeat([]byte{0x11}, DEKBytes)
	key, _ := NewKEK("export-artifact.v1", material)
	material[0] = 0
	keyring, _ := NewEnvelopeKeyring(key, nil,
		bytes.NewReader(bytes.Repeat([]byte{0x44}, DEKBytes+NonceBytes)))
	_, envelope, err := keyring.Generate(context.Background(), envelopeIdentity())
	if err != nil {
		t.Fatal(err)
	}
	copyEnvelope := envelope.OwnedCopy()
	envelope.Ciphertext[0] ^= 1
	if copyEnvelope.Ciphertext[0] == envelope.Ciphertext[0] {
		t.Fatal("key envelope aliases caller storage")
	}
	if fmt.Sprint(key, copyEnvelope) !=
		"[REDACTED FCART1 KEK] [REDACTED FCART1 KEY ENVELOPE]" {
		t.Fatal("key material formatting was not redacted")
	}
}

func envelopeIdentity() KeyIdentity {
	return KeyIdentity{TenantID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ArtifactID:        uuid.MustParse("019b22de-8a00-7000-8000-000000000001"),
		AttemptGeneration: 2}
}
