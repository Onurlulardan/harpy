package postgres

import "testing"

func TestNewStoreBuildsImmutableCanonicalMetadata(t *testing.T) {
	selected := validSelectedMigrations()
	store, err := NewStore(testRunID, selected)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	if store.Tablename() != historyTableName || len(store.metadataByVersion) != len(selected) {
		t.Fatal("store lost exact table name or selected metadata")
	}
	version := selected[0].Version
	selected[0].Owner = "tampered"
	metadata := store.metadataByVersion[version]
	if metadata.owner != "platform_migrator" || metadata.runID != testRunID {
		t.Fatal("store retained mutable constructor input")
	}
	if _, exists := store.metadataByVersion[0]; exists {
		t.Fatal("version zero anchor was accepted as migration metadata")
	}
}

func TestNewStoreRejectsInvalidRunIdentityAndEmptyMetadata(t *testing.T) {
	if _, err := NewStore("11111111-1111-0111-8111-111111111111", validSelectedMigrations()); err == nil {
		t.Fatal("non-versioned run UUID accepted")
	}
	if _, err := NewStore(testRunID, nil); err == nil {
		t.Fatal("empty migration metadata accepted")
	}
}
