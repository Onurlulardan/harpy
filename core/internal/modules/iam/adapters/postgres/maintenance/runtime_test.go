package maintenance

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestRuntimeCloseWipesOwnedDirectoryKeyMaterial(t *testing.T) {
	key := DirectoryKey{
		ID:       base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{22}, 12)),
		Material: [32]byte{1, 2, 3}, Active: true,
	}
	runtime, err := NewRuntime([]DirectoryKey{key})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	owned := runtime.keys
	if err := runtime.Close(); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if owned[0].ID != "" || owned[0].Material != [32]byte{} || owned[0].Active {
		t.Fatalf("closed key was not wiped: %#v", owned[0])
	}
	if runtime.usable() {
		t.Fatal("closed runtime remained usable")
	}
}
