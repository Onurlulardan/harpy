package keyring

import (
	"bytes"
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/keywrap"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestRingCreatesContextBoundRandomDataKey(t *testing.T) {
	key := [32]byte{7}
	randomness := bytes.NewReader(make([]byte, 96))
	ring, err := New([]Release{{ID: "documents-2026-07", Current: true, Key: key}},
		randomness)
	if err != nil {
		t.Fatal(err)
	}
	tenant := domain.ID{1}
	document := domain.ID{6: 0x70, 8: 0x80, 15: 1}
	version := domain.ID{6: 0x70, 8: 0x80, 15: 2}
	dataKey, err := ring.CreateDataKey(context.Background(), tenant, document, version)
	if err != nil || dataKey.KeyID != "documents-2026-07" {
		t.Fatalf("key=%+v err=%v", dataKey, err)
	}
	opened, err := keywrap.Unwrap(dataKey.Wrapped, key[:],
		wrapContext(tenant, document, version, dataKey.KeyID))
	if err != nil || !bytes.Equal(opened, dataKey.Plaintext[:]) {
		t.Fatalf("opened=%x err=%v", opened, err)
	}
}

func TestRingRequiresExactlyOneCurrentRelease(t *testing.T) {
	key := [32]byte{1}
	if _, err := New([]Release{{ID: "one", Key: key}}, bytes.NewReader(nil)); err == nil {
		t.Fatal("ring without current key was accepted")
	}
}
