package migrationsupersession

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectReplacesBothOriginsAtOneVersion(t *testing.T) {
	catalog, sources := decodeFixture(t)
	selected, err := Select(catalog, sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 {
		t.Fatalf("selected documents = %d", len(selected))
	}
	for _, document := range selected {
		if !strings.Contains(document.Path, "/supersessions/") ||
			!bytes.Contains(document.Content, []byte("bigint")) {
			t.Fatalf("old source remained selected: %s", document.Path)
		}
	}
}

func TestSelectRejectsTamperedContentAndDeclaredHash(t *testing.T) {
	catalog, sources := decodeFixture(t)
	sources[0].Content = append(sources[0].Content, ' ')
	if _, err := Select(catalog, sources); err == nil {
		t.Fatal("tampered old source was accepted")
	}
	encoded, sources := fixture(t)
	encoded = bytes.Replace(encoded, []byte(`"old_sha256":"`),
		[]byte(`"old_sha256":"0`), 1)
	if _, err := DecodeCatalog(encoded); err == nil {
		t.Fatal("invalid declared SHA-256 was accepted")
	}
}

func TestSelectRejectsMissingAndOrphanReplacement(t *testing.T) {
	catalog, sources := decodeFixture(t)
	if _, err := Select(catalog, sources[:len(sources)-1]); err == nil {
		t.Fatal("missing replacement was accepted")
	}
	sources = append(sources, Source{
		Path:    fixtureDirectory + "/supersessions/fresh/20260721010001_orphan_v1.fresh.sql",
		Content: []byte("-- +goose Up\nSELECT 2;\n"),
	})
	if _, err := Select(catalog, sources); err == nil {
		t.Fatal("orphan replacement was accepted")
	}
}

func TestRejectReservedPathsRequiresCatalog(t *testing.T) {
	_, sources := fixture(t)
	if err := RejectReservedPaths(fixtureDirectory, sources); err == nil {
		t.Fatal("reserved replacement path without catalog was accepted")
	}
}
