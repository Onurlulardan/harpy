package migrationsupersession

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectV2ChainLeavesOnlyTerminalReplacement(t *testing.T) {
	catalog, sources := decodeChainFixture(t)
	selected, err := Select(catalog, sources)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ContractVersion() != ContractVersionV2 || len(selected) != 2 {
		t.Fatalf("v2 chain selection differs: %s, %d",
			catalog.ContractVersion(), len(selected))
	}
	for _, document := range selected {
		if !strings.Contains(document.Path, "/supersessions/v2/") ||
			!bytes.Contains(document.Content, []byte("numeric")) {
			t.Fatalf("non-terminal source remained selected: %s", document.Path)
		}
		lineage := catalog.Lineage(document.Path)
		if len(lineage) != 2 ||
			lineage[0].ReplacementPath() != lineage[1].OldPath() ||
			lineage[0].ReplacementSHA256() != lineage[1].OldSHA256() {
			t.Fatalf("terminal lineage differs: %#v", lineage)
		}
	}
}

func TestDecodeRejectsIncompleteOrBrokenV2Chain(t *testing.T) {
	encoded, _ := chainFixture(t)
	var decoded catalogDocument
	decodeCatalogDocument(t, encoded, &decoded)
	decoded.Entries = decoded.Entries[:len(decoded.Entries)-1]
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("incomplete v2 chain was accepted")
	}

	decodeCatalogDocument(t, encoded, &decoded)
	for index := range decoded.Entries {
		if strings.Contains(decoded.Entries[index].OldPath, "/supersessions/fresh/") {
			decoded.Entries[index].OldSHA256 = strings.Repeat("f", 64)
		}
	}
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("broken v2 chain hash link was accepted")
	}
}

func TestDecodeV1RejectsSecondLevelReplacement(t *testing.T) {
	encoded, _ := chainFixture(t)
	var decoded catalogDocument
	decodeCatalogDocument(t, encoded, &decoded)
	decoded.ContractVersion = ContractVersionV1
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("v1 catalog accepted a second-level replacement")
	}
}
