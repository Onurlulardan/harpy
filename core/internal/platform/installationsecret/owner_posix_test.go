//go:build !windows

package installationsecret

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadRejectsBroadPermissionsAndSymlink(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "initial-password")
	if err := Publish(path, canonicalWire(0x5d)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := Read(path, func([]byte) error { return nil }); err != ErrFenced {
		t.Fatalf("broad permission err=%v", err)
	}
	link := filepath.Join(directory, "linked-password")
	if err := os.Symlink(path, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := Read(link, func([]byte) error { return nil }); err != ErrFenced {
		t.Fatalf("symlink err=%v", err)
	}
}
