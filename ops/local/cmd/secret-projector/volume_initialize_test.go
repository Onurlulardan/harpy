package main

import (
	"os"
	"testing"
)

func TestInitializeStateVolumePreservesContent(t *testing.T) {
	root := t.TempDir()
	path := root + string(os.PathSeparator) + "receipt"
	if err := os.WriteFile(path, []byte("persisted"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := targetSpec{name: "release", uid: runtimeUserID,
		gid: runtimeUserID, dir: 0o700}
	if err := initializeStateVolume(root, target, noOwnershipChange); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "persisted" {
		t.Fatal("state initializer changed persisted content")
	}
}
