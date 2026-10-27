package api

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

const validCatalogHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestCatalogHashCanonicalRoundTripAndOwnedBytes(t *testing.T) {
	hash, err := NewCatalogHash(validCatalogHash)
	if err != nil || hash.Validate() != nil {
		t.Fatalf("valid catalog hash rejected: %v", err)
	}
	hexValue, err := hash.Hex()
	if err != nil || hexValue != validCatalogHash {
		t.Fatalf("hex round trip = %q, %v", hexValue, err)
	}
	want, _ := hex.DecodeString(validCatalogHash)
	actual, err := hash.Bytes()
	if err != nil || !bytes.Equal(actual, want) {
		t.Fatalf("bytes round trip failed: %x, %v", actual, err)
	}
	actual[0] ^= 0xff
	again, err := hash.Bytes()
	if err != nil || !bytes.Equal(again, want) {
		t.Fatal("catalog hash exposed mutable byte storage")
	}
}

func TestCatalogHashRejectsNonCanonicalOrZeroInput(t *testing.T) {
	for _, value := range []string{
		"", validCatalogHash[:63], validCatalogHash + "0",
		strings.ToUpper(validCatalogHash), strings.Repeat("0", 64),
		strings.Repeat("g", 64),
	} {
		if _, err := NewCatalogHash(value); err == nil {
			t.Fatalf("invalid catalog hash accepted: %q", value)
		} else if strings.Contains(err.Error(), value) && value != "" {
			t.Fatalf("catalog hash error leaked input: %v", err)
		}
	}
}

func TestCatalogHashZeroAndTamperFailClosed(t *testing.T) {
	for _, hash := range []CatalogHash{
		{},
		{parsed: true},
		{digest: [32]byte{1}},
	} {
		if hash.Validate() == nil {
			t.Fatal("invalid catalog hash validated")
		}
		if _, err := hash.Hex(); err == nil {
			t.Fatal("invalid catalog hash exposed hex")
		}
		if value, err := hash.Bytes(); err == nil || value != nil {
			t.Fatal("invalid catalog hash exposed bytes")
		}
	}
}
