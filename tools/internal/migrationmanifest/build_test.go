package migrationmanifest

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestBuildPairsOriginsSortsEntriesAndHashesRawSQL(t *testing.T) {
	group := testSourceGroup(t)
	second := testSourceGroup(t)
	for index, document := range second.Documents {
		second.Documents[index] = testDocument(
			t, testOwner,
			strings.Replace(document.Path(), testVersion, "20260713010100", 1),
			document.Content(),
		)
	}
	group.Documents = append(group.Documents, second.Documents...)
	manifest, err := Build(testStorageInventory(group), []SourceGroup{group})
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Entries) != 2 || manifest.Entries[0].Version != 20260713010000 ||
		manifest.Entries[1].Version != 20260713010100 {
		t.Fatalf("generated entries are not globally sorted: %#v", manifest.Entries)
	}
	first := manifest.Entries[0]
	if first.Variants.Fresh == nil || first.Variants.V1Adopted == nil ||
		first.Variants.Fresh.SHA256 == first.Variants.V1Adopted.SHA256 {
		t.Fatal("origin variants did not retain exact raw SQL identity")
	}
	canonical, digest, err := Encode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(canonical)
	if err != nil || verified.DigestSHA256() != digest {
		t.Fatalf("encoded manifest did not verify: %v", err)
	}
}

func TestBuildRejectsIncompleteAndConflictingSources(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SourceGroup)
	}{
		{"missing origin", func(group *SourceGroup) { group.Documents = group.Documents[:1] }},
		{"duplicate origin", func(group *SourceGroup) {
			group.Documents = append(group.Documents, group.Documents[0])
		}},
		{"owner drift", func(group *SourceGroup) {
			original := group.Documents[0]
			group.Documents[0] = testDocument(
				t, "platform_cutover", original.Path(), original.Content(),
			)
		}},
		{"logical name drift", func(group *SourceGroup) {
			original := group.Documents[0]
			group.Documents[0] = testDocument(
				t, testOwner,
				strings.Replace(original.Path(), "create_foundation", "create_other", 1),
				original.Content(),
			)
		}},
		{"transaction drift", func(group *SourceGroup) {
			original := group.Documents[0]
			group.Documents[0] = testDocument(
				t, testOwner, original.Path(),
				[]byte("-- +goose NO TRANSACTION\n-- +goose Up\nSELECT 2;\n"),
			)
		}},
		{"invalid scope", func(group *SourceGroup) { group.Scope = "shared" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			group := testSourceGroup(t)
			test.mutate(&group)
			if _, err := Build(testStorageInventory(group), []SourceGroup{group}); err == nil {
				t.Fatal("invalid migration sources were accepted")
			}
		})
	}
}

func TestBuildRejectsEmptySourcesAndDuplicateVersionAcrossOwners(t *testing.T) {
	if _, err := Build(StorageInventory{}, nil); err == nil {
		t.Fatal("empty migration sources were accepted")
	}
	first := testSourceGroup(t)
	second := testSourceGroup(t)
	second.Owner = "platform_cutover"
	for index, document := range second.Documents {
		second.Documents[index] = testDocument(
			t, second.Owner, document.Path(), document.Content(),
		)
	}
	if _, err := Build(testStorageInventory(first), []SourceGroup{first, second}); err == nil {
		t.Fatal("same migration version across owners was accepted")
	}
}
