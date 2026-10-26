package migrationsupersession

import (
	"bytes"
	"testing"
)

func TestDecodeAcceptsCanonicalPairedCatalog(t *testing.T) {
	encoded, _ := fixture(t)
	catalog, err := DecodeCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Owner() != "platform_example" || catalog.Scope() != "platform" ||
		catalog.SourceDirectory() != fixtureDirectory || len(catalog.Entries()) != 2 {
		t.Fatalf("decoded catalog differs: %#v", catalog)
	}
}

func TestDecodeRejectsNonCanonicalAndPathMismatch(t *testing.T) {
	encoded, _ := fixture(t)
	if _, err := DecodeCatalog(append([]byte("\n"), encoded...)); err == nil {
		t.Fatal("non-canonical catalog was accepted")
	}
	mismatched := bytes.Replace(
		encoded,
		[]byte("/supersessions/fresh/20260721010000"),
		[]byte("/fresh/20260721010000"),
		1,
	)
	if _, err := DecodeCatalog(mismatched); err == nil {
		t.Fatal("replacement path mismatch was accepted")
	}
}

func TestDecodeRejectsDuplicateAndUnpairedEntries(t *testing.T) {
	encoded, _ := fixture(t)
	var decoded catalogDocument
	decodeCatalogDocument(t, encoded, &decoded)
	decoded.Entries = append(decoded.Entries, decoded.Entries[len(decoded.Entries)-1])
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("duplicate entry was accepted")
	}
	decoded.Entries = decoded.Entries[:1]
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("unpaired entry was accepted")
	}
}
