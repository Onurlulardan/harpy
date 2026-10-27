package componentowner

import "testing"

func TestDirectoryRejectsIncompleteOwners(t *testing.T) {
	if _, err := NewDirectory(map[string]Owner{}); err == nil {
		t.Fatal("incomplete owner directory accepted")
	}
}

func TestDirectoryRejectsUnknownOwner(t *testing.T) {
	var directory *StaticDirectory
	if _, err := directory.Resolve("unknown"); err == nil {
		t.Fatal("unknown owner resolved")
	}
}
