package query

import "testing"

func TestPublishedVersionAndRetiredRootAreImmutableTransitions(t *testing.T) {
	queryID, _ := NewID(v7(1))
	versionID, _ := NewVersionID(v7(2))
	root, err := Rehydrate(queryID, LifecycleActive, 4, 2, versionID)
	if err != nil {
		t.Fatal(err)
	}
	retired, err := root.Retire(4)
	if err != nil || retired.RowRevision() != 5 || retired.ReferenceEpoch() != 3 ||
		retired.Lifecycle() != LifecycleRetired || retired.CurrentDraft().Valid() {
		t.Fatalf("retire result = %#v, %v", retired, err)
	}
	digest := [32]byte{1}
	version, err := RehydrateVersion(versionID, queryID, 1, 3, VersionDraft, digest, digest)
	if err != nil {
		t.Fatal(err)
	}
	published, err := version.Publish(3)
	if err != nil || published.State() != VersionPublished || published.Revision() != 4 {
		t.Fatalf("publish result = %#v, %v", published, err)
	}
	if _, err := published.Publish(4); err != ErrConflict {
		t.Fatalf("republish error = %v", err)
	}
}

func v7(seed byte) [16]byte {
	value := [16]byte{0, 0, 0, 0, 0, 0, 0x70, seed, 0x80}
	value[15] = seed
	return value
}
