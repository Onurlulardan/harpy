package cipher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"testing"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestRuntimeEncryptsExactFCDOC1Stream(t *testing.T) {
	runtime, err := New(bytes.NewReader(make([]byte, 128)))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("document")
	metadata := uploadapp.CipherMetadata{TenantID: domain.ID{1},
		DocumentID: domain.ID{6: 0x70, 8: 0x80, 15: 1},
		VersionID:  domain.ID{6: 0x70, 8: 0x80, 15: 2}, Generation: 1,
		PlaintextBytes: int64(len(plain))}
	var encrypted bytes.Buffer
	receipt, err := runtime.Encrypt(context.Background(), bytes.NewReader(plain),
		&encrypted, [32]byte{1}, metadata)
	if err != nil || receipt.PlaintextSHA256 != sha256.Sum256(plain) ||
		receipt.CiphertextBytes != int64(encrypted.Len()) {
		t.Fatalf("receipt=%+v bytes=%d err=%v", receipt, encrypted.Len(), err)
	}
}
