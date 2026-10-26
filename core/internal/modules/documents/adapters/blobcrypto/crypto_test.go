package blobcrypto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestFCDOC1RoundTripPinsIdentityAndHashes(t *testing.T) {
	metadata := fixtureMetadata(t, int64(len("secret document")))
	key := bytes.Repeat([]byte{0x41}, 32)
	randomness := bytes.NewReader(bytes.Repeat([]byte{0x24}, 12))
	var encrypted bytes.Buffer
	encryptReceipt, err := Encrypt(
		bytes.NewBufferString("secret document"), &encrypted, key, metadata, randomness,
	)
	if err != nil {
		t.Fatal(err)
	}
	var opened bytes.Buffer
	decryptReceipt, err := Decrypt(bytes.NewReader(encrypted.Bytes()), &opened, key, metadata)
	if err != nil || opened.String() != "secret document" {
		t.Fatalf("opened=%q receipt=%+v err=%v", opened.String(), decryptReceipt, err)
	}
	if decryptReceipt != encryptReceipt || decryptReceipt.Metadata != metadata {
		t.Fatalf("receipt mismatch: %+v / %+v", encryptReceipt, decryptReceipt)
	}
}

func TestFCDOC1RejectsTamperAndIdentitySubstitution(t *testing.T) {
	metadata := fixtureMetadata(t, 7)
	key := bytes.Repeat([]byte{0x41}, 32)
	var encrypted bytes.Buffer
	_, err := Encrypt(bytes.NewBufferString("content"), &encrypted, key, metadata,
		bytes.NewReader(bytes.Repeat([]byte{0x25}, 12)))
	if err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Clone(encrypted.Bytes())
	tampered[len(tampered)-1] ^= 0xff
	if _, err = Decrypt(bytes.NewReader(tampered), &bytes.Buffer{}, key, metadata); !errors.Is(err, ErrContentTampered) {
		t.Fatalf("tamper error = %v", err)
	}
	other := metadata
	other.Generation++
	if _, err = Decrypt(bytes.NewReader(encrypted.Bytes()), &bytes.Buffer{}, key, other); !errors.Is(err, ErrIdentityMismatch) {
		t.Fatalf("identity error = %v", err)
	}
}

func fixtureMetadata(t *testing.T, size int64) Metadata {
	t.Helper()
	parse := func(value string) domain.ID {
		id, err := domain.ParseID(value)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	return Metadata{
		TenantID:   parse("11111111-1111-4111-8111-111111111111"),
		DocumentID: parse("018f0000-0000-7000-8000-000000000002"),
		VersionID:  parse("018f0000-0000-7000-8000-000000000003"),
		Generation: 4, PlaintextBytes: size,
	}
}
