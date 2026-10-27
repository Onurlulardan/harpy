package config

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestZeroCursorKeyringFailsClosed(t *testing.T) {
	var zero CursorKeyring
	if err := zero.Validate(); err == nil {
		t.Fatal("zero cursor keyring validated")
	}
	if _, err := zero.Current(); err == nil {
		t.Fatal("zero cursor keyring exposed current key")
	}
	if _, err := zero.Previous(); err == nil {
		t.Fatal("zero cursor keyring exposed previous key")
	}
}

func TestZeroCursorKeyFailsClosed(t *testing.T) {
	var zero CursorKey
	if _, err := zero.KID(); err == nil {
		t.Fatal("zero cursor key exposed KID")
	}
	if _, err := zero.AES256Key(); err == nil {
		t.Fatal("zero cursor key exposed material")
	}
}

func TestCopiedCursorKeyringPreservesValidatedZeroMaterial(t *testing.T) {
	document := cursorTestJSON(
		"current-zero", cursorTestKey(0), "previous-1", cursorTestKey(1),
	)
	original, err := parseCursorKeyring(Secret{value: document})
	if err != nil {
		t.Fatal(err)
	}
	copied := original
	if err := copied.Validate(); err != nil {
		t.Fatalf("copied keyring became invalid: %v", err)
	}
	current, err := copied.Current()
	if err != nil {
		t.Fatal(err)
	}
	material, err := current.AES256Key()
	if err != nil {
		t.Fatal(err)
	}
	if material != [cursorKeySize]byte{} {
		t.Fatal("valid all-zero key material changed")
	}
}

func TestInvalidCursorKeyErrorsDoNotLeakFields(t *testing.T) {
	material := [cursorKeySize]byte{9}
	invalid := CursorKey{kid: "must-not-leak", material: material}
	invalidRing := CursorKeyring{current: invalid, previous: invalid, valid: true}
	for _, access := range []func() error{
		invalidRing.Validate,
		func() error { _, err := invalidRing.Current(); return err },
		func() error { _, err := invalidRing.Previous(); return err },
		func() error { _, err := invalid.KID(); return err },
		func() error { _, err := invalid.AES256Key(); return err },
	} {
		err := access()
		if err == nil || cursorErrorContainsField(err, invalid.kid, material) {
			t.Fatalf("invalid key access returned unsafe error: %v", err)
		}
	}
}

func cursorErrorContainsField(err error, kid string, material [cursorKeySize]byte) bool {
	message := err.Error()
	return strings.Contains(message, kid) ||
		strings.Contains(message, fmt.Sprint(material)) ||
		strings.Contains(message, base64.RawURLEncoding.EncodeToString(material[:]))
}
