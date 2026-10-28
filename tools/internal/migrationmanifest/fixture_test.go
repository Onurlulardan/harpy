package migrationmanifest

import (
	"testing"
)

const (
	testOwner   = "platform_migrator"
	testVersion = "20260713010000"
	testRoot    = "core/internal/platform/migrator/migrations/"
)

func testSourceGroup(t *testing.T) SourceGroup {
	t.Helper()
	return SourceGroup{
		Owner: testOwner, Scope: "platform", Phase: "foundation",
		TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
		VersionFrom: 20260713010000, VersionTo: 20260713010100,
		Directory: testRoot[:len(testRoot)-1],
		Documents: []SourceDocument{
			testDocument(t, testOwner, testRoot+"v1_adopted/"+testVersion+
				"_create_foundation.v1_adopted.sql", []byte("-- +goose Up\nSELECT 2;\n")),
			testDocument(t, testOwner, testRoot+"fresh/"+testVersion+
				"_create_foundation.fresh.sql", []byte("-- +goose Up\nSELECT 1;\n")),
		},
	}
}

func testStorageInventory(group SourceGroup) StorageInventory {
	return StorageInventory{
		claims: []StorageClaim{{
			Owner: group.Owner, Directory: group.Directory, Scope: "platform",
		}},
		streams: []MigrationStream{{
			Owner: testOwner, Directory: group.Directory, Scope: "platform",
			VersionFrom: group.VersionFrom, VersionTo: group.VersionTo,
			Phase: "foundation", TargetDatabaseGeneration: 1,
			NextMinimumRuntimeGeneration: 1,
		}},
	}
}

func testDocument(
	t *testing.T,
	owner string,
	path string,
	content []byte,
) SourceDocument {
	t.Helper()
	document, err := NewSourceDocument(owner, path, content)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
