package keywrap

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestWrapPinsIdentityRejectsTamperAndRewraps(t *testing.T) {
	dek := bytes.Repeat([]byte{0x11}, 32)
	oldKEK := bytes.Repeat([]byte{0x22}, 32)
	newKEK := bytes.Repeat([]byte{0x33}, 32)
	oldContext := fixtureContext(t, "documents-kek-1")
	wrapped, err := Wrap(dek, oldKEK, oldContext,
		bytes.NewReader(bytes.Repeat([]byte{0x44}, 12)))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := Unwrap(wrapped, oldKEK, oldContext)
	if err != nil || !bytes.Equal(opened, dek) {
		t.Fatalf("opened=%x err=%v", opened, err)
	}
	wrong := oldContext
	wrong.KeyID = "documents-kek-2"
	if _, err = Unwrap(wrapped, oldKEK, wrong); !errors.Is(err, ErrInvalidWrap) {
		t.Fatalf("identity error = %v", err)
	}
	tampered := bytes.Clone(wrapped)
	tampered[len(tampered)-1] ^= 0xff
	if _, err = Unwrap(tampered, oldKEK, oldContext); !errors.Is(err, ErrKeyTampered) {
		t.Fatalf("tamper error = %v", err)
	}
	newContext := oldContext
	newContext.KeyID = "documents-kek-2"
	rewrapped, err := Rewrap(wrapped, oldKEK, newKEK, oldContext, newContext,
		bytes.NewReader(bytes.Repeat([]byte{0x55}, 12)))
	if err != nil {
		t.Fatal(err)
	}
	opened, err = Unwrap(rewrapped, newKEK, newContext)
	if err != nil || !bytes.Equal(opened, dek) {
		t.Fatalf("rewrapped=%x err=%v", opened, err)
	}
}

func fixtureContext(t *testing.T, keyID string) Context {
	t.Helper()
	parse := func(value string) domain.ID {
		id, err := domain.ParseID(value)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	return Context{
		TenantID:   parse("11111111-1111-4111-8111-111111111111"),
		DocumentID: parse("018f0000-0000-7000-8000-000000000002"),
		VersionID:  parse("018f0000-0000-7000-8000-000000000003"), KeyID: keyID,
	}
}
