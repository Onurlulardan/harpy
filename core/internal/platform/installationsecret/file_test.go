package installationsecret

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOwnerOnlyFilePublishReadRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-password")
	wire := canonicalWire(0x2a)
	if err := Publish(path, wire); err != nil {
		t.Fatal(err)
	}
	if err := Publish(path, canonicalWire(0x3b)); err != ErrExists {
		t.Fatalf("overwrite err=%v", err)
	}
	if err := Read(path, func(value []byte) error {
		if !bytes.Equal(value, wire) {
			t.Fatal("wire changed")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("secret remains: %v", err)
	}
}

func TestOwnerOnlyFileRejectsMalformedAndConsumerError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-password")
	if err := Publish(path, []byte("invalid")); err != ErrInvalid {
		t.Fatalf("malformed err=%v", err)
	}
	if err := Publish(path, canonicalWire(0x4c)); err != nil {
		t.Fatal(err)
	}
	want := errors.New("display failed")
	if err := Read(path, func([]byte) error { return want }); !errors.Is(err, want) {
		t.Fatalf("consumer err=%v", err)
	}
}

func canonicalWire(value byte) []byte {
	raw := bytes.Repeat([]byte{value}, rawPasswordBytes)
	wire := make([]byte, wirePasswordBytes)
	base64.RawURLEncoding.Encode(wire, raw)
	clear(raw)
	return wire
}
