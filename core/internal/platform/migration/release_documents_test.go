package migration

import (
	"crypto/sha256"
	"fmt"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestSelectAuthorizedDocumentsNarrowsToRelease(t *testing.T) {
	target := testReleaseDocument(t, "target.sql", "-- +goose Up\nSELECT 1;\n")
	future := testReleaseDocument(t, "future.sql", "-- +goose Up\nSELECT 2;\n")
	registry := testDocumentRegistry(t, target, future)
	manifest := testDocumentManifest(target)
	selected, err := registry.selectAuthorizedDocuments(manifest, []Document{target, future})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].Path() != target.Path() {
		t.Fatal("release document selection differs")
	}
	unauthorized := testReleaseDocument(t, "unknown.sql", "-- +goose Up\nSELECT 3;\n")
	if _, err := registry.selectAuthorizedDocuments(
		manifest, []Document{target, future, unauthorized},
	); err == nil {
		t.Fatal("unauthorized embedded migration source was accepted")
	}
}

func testReleaseDocument(t *testing.T, path, source string) Document {
	t.Helper()
	document, err := NewDocument("platform_migrator", "platform", path, []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func testDocumentRegistry(t *testing.T, documents ...Document) *ReleaseRegistry {
	t.Helper()
	registry := &ReleaseRegistry{documentContracts: map[string]map[expectedDocument]struct{}{}}
	for _, document := range documents {
		metadata, err := InspectDocument(document)
		if err != nil {
			t.Fatal(err)
		}
		contract := expectedDocument{owner: document.Owner(), scope: document.Scope(),
			transactionMode: metadata.TransactionMode, sha256: metadata.SHA256}
		registry.documentContracts[document.Path()] = map[expectedDocument]struct{}{contract: {}}
	}
	return registry
}

func testDocumentManifest(document Document) postgresv1.MigrationManifest {
	digest := fmt.Sprintf("%x", sha256.Sum256(document.Content()))
	return postgresv1.MigrationManifest{Entries: []postgresv1.MigrationEntry{{
		Owner: document.Owner(), Scope: document.Scope(), TransactionMode: "transactional",
		Variants: postgresv1.MigrationVariants{
			Fresh: &postgresv1.MigrationVariant{Path: document.Path(), SHA256: digest},
		},
	}}}
}
