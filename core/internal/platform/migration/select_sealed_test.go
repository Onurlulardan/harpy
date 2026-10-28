package migration

import (
	"strings"
	"testing"
)

func TestSealedReleaseSkipsUnrelatedSQL(t *testing.T) {
	target := testReleaseDocument(t, "target.sql", "-- +goose Up\nSELECT 1;\n")
	later := testReleaseDocument(t, "later.sql", "-- +goose Up\nSELECT 2;\n")
	registry := sealedDocumentRegistry(t, target)
	selected, err := registry.selectAuthorizedDocuments(
		testDocumentManifest(target), []Document{target, later},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].Path() != target.Path() {
		t.Fatal("sealed selection included a file the release does not name")
	}
}

func TestSealedReleaseRejectsWrongHash(t *testing.T) {
	target := testReleaseDocument(t, "target.sql", "-- +goose Up\nSELECT 1;\n")
	mutated := testReleaseDocument(t, "target.sql", "-- +goose Up\nSELECT 9;\n")
	registry := sealedDocumentRegistry(t, target)
	_, err := registry.selectAuthorizedDocuments(
		testDocumentManifest(target), []Document{mutated},
	)
	if err == nil || !strings.Contains(err.Error(), "not release-authorized") {
		t.Fatalf("wrong sealed hash error = %v", err)
	}
}

func TestSealedReleaseRejectsMissingNamedFile(t *testing.T) {
	target := testReleaseDocument(t, "target.sql", "-- +goose Up\nSELECT 1;\n")
	later := testReleaseDocument(t, "later.sql", "-- +goose Up\nSELECT 2;\n")
	registry := sealedDocumentRegistry(t, target)
	if _, err := registry.selectAuthorizedDocuments(
		testDocumentManifest(target), []Document{later},
	); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("missing sealed file error = %v", err)
	}
}

func sealedDocumentRegistry(t *testing.T, documents ...Document) *ReleaseRegistry {
	t.Helper()
	registry := testDocumentRegistry(t, documents...)
	registry.known = map[string]struct{}{"sealed": {}}
	return registry
}
