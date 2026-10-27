package installationsecret

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestWP10OwnerOnlyCredentialIsDisplayedAndDeletedOnce(t *testing.T) {
	password, err := GenerateTemporaryPassword()
	if err != nil {
		t.Fatal(err)
	}
	defer password.Destroy()
	var plaintext []byte
	if err = password.Consume(func(value []byte) error {
		plaintext = append([]byte(nil), value...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	defer clear(plaintext)
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(string(plaintext))
	if err != nil || len(decoded) != 24 {
		t.Fatal("temporary credential does not contain exactly 24 random bytes")
	}
	clear(decoded)
	path := filepath.Join(t.TempDir(), "initial-password")
	if err = Publish(path, plaintext); err != nil {
		t.Fatal(err)
	}
	var displayed []byte
	if err = Read(path, func(value []byte) error {
		displayed = append([]byte(nil), value...)
		return nil
	}); err != nil || !bytes.Equal(displayed, plaintext) {
		t.Fatal("owner-only credential display failed")
	}
	clear(displayed)
	if err = Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temporary credential still exists: %v", err)
	}
}
