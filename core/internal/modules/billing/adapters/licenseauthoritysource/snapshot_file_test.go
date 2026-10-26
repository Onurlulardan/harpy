package licenseauthoritysource

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSnapshotFileReadsBoundedRegularOwnerOnlyArtifact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	document := []byte(`{"schema":"harpy-license-envelope.v2"}`)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := NewSnapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	read, err := source.IssueSnapshot(context.Background(), connectedRequestFixture())
	if err != nil || !bytes.Equal(read, document) {
		t.Fatalf("snapshot read failed: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err = os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err = source.IssueSnapshot(
			context.Background(), connectedRequestFixture(),
		); err == nil {
			t.Fatal("non-owner-only snapshot accepted")
		}
	}
}

func TestSnapshotFileRejectsRelativeAndDirectoryPaths(t *testing.T) {
	if _, err := NewSnapshotFile("snapshot.json"); err == nil {
		t.Fatal("relative snapshot path accepted")
	}
	directory := t.TempDir()
	source, err := NewSnapshotFile(directory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = source.IssueSnapshot(
		context.Background(), connectedRequestFixture(),
	); err == nil {
		t.Fatal("directory accepted as snapshot file")
	}
}
