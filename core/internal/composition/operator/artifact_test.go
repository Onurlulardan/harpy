package operator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadRegularFileIsBoundedAndReturnsPrivateCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential")
	if err := os.WriteFile(path, []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	content, err := readRegularFile(path, 16, true)
	if err != nil || string(removeOneTerminalNewline(content)) != "secret" {
		t.Fatalf("content=%q err=%v", content, err)
	}
	clear(content)
	if _, err := readRegularFile(path, 3, true); err == nil {
		t.Fatal("oversized credential was accepted")
	}
}

func TestReadRegularFileRejectsSymlink(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(filepath.Dir(target), "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := readRegularFile(link, 16, true); err == nil {
		t.Fatal("symlink credential was accepted")
	}
}
