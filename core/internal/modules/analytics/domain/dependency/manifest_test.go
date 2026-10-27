package dependency

import "testing"

func TestManifestIsOrderIndependentAndRevisionBoundFailsClosed(t *testing.T) {
	digest := [32]byte{1}
	left, err := NewEntry("iam", "directory", id(1), id(2), 1, 1, 1,
		"iam.directory.revision.v1", digest)
	if err != nil {
		t.Fatal(err)
	}
	right, rightErr := NewEntry("iam", "role", id(3), id(4), 2, 1, 3,
		"iam.role.revision.v1", digest)
	if rightErr != nil {
		t.Fatal(rightErr)
	}
	first, err := NewManifest([]Entry{left, right}, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewManifest([]Entry{right, left}, true)
	if err != nil || !first.Equal(second) {
		t.Fatalf("manifest parity = %v, %v", first.Equal(second), err)
	}
	withoutRevision, revisionErr := NewEntry("iam", "directory", id(5), id(6), 1, 1, 1, "", digest)
	if revisionErr != nil {
		t.Fatal(revisionErr)
	}
	if _, err := NewManifest([]Entry{withoutRevision}, true); err != ErrInvalid {
		t.Fatalf("revision-bound manifest error = %v", err)
	}
	if _, err := NewEntry("work", "board", id(7), id(8), 1, 1, 1, "x", digest); err != ErrInvalid {
		t.Fatalf("later-owner error = %v", err)
	}
}

func id(seed byte) [16]byte { value := [16]byte{}; value[15] = seed; return value }
