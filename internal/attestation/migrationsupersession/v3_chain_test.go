package migrationsupersession

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func v3ChainFixture(t *testing.T) ([]byte, []Source) {
	t.Helper()
	contents := [][]byte{
		[]byte("-- +goose Up\nSELECT 1;\n"),
		[]byte("-- +goose Up\nSELECT 1::bigint;\n"),
		[]byte("-- +goose Up\nSELECT 1::numeric;\n"),
		[]byte("-- +goose Up\nSELECT 1::double precision;\n"),
	}
	hashes := make([]string, len(contents))
	for index, content := range contents {
		hashes[index] = fmt.Sprintf("%x", sha256.Sum256(content))
	}
	document := catalogDocument{
		ContractVersion: ContractVersionV3,
		Owner:           "platform_example",
		Scope:           "platform",
		SourceDirectory: fixtureDirectory,
	}
	var sources []Source
	for _, origin := range []string{"fresh", "v1_adopted"} {
		base := "20260721010000_compile_pg18_v1." + origin + ".sql"
		paths := []string{
			fixtureDirectory + "/" + origin + "/" + base,
			fixtureDirectory + "/supersessions/" + origin + "/" + base,
			fixtureDirectory + "/supersessions/v2/" + origin + "/" + base,
			fixtureDirectory + "/supersessions/v3/" + origin + "/" + base,
		}
		for level := 0; level < 3; level++ {
			document.Entries = append(document.Entries, entryDocument{
				OldPath: paths[level], OldSHA256: hashes[level],
				ReplacementPath: paths[level+1], ReplacementSHA256: hashes[level+1],
			})
		}
		for index, sourcePath := range paths {
			sources = append(sources, Source{Path: sourcePath, Content: contents[index]})
		}
	}
	sort.Slice(document.Entries, func(left, right int) bool {
		return document.Entries[left].OldPath < document.Entries[right].OldPath
	})
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return append(canonical, '\n'), sources
}

func TestSelectV3ChainLeavesOnlyTerminalReplacement(t *testing.T) {
	encoded, sources := v3ChainFixture(t)
	catalog, err := DecodeCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := Select(catalog, sources)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ContractVersion() != ContractVersionV3 || len(selected) != 2 {
		t.Fatalf("v3 chain selection differs: %s, %d",
			catalog.ContractVersion(), len(selected))
	}
	for _, document := range selected {
		if !strings.Contains(document.Path, "/supersessions/v3/") ||
			!bytes.Contains(document.Content, []byte("double precision")) {
			t.Fatalf("non-terminal source remained selected: %s", document.Path)
		}
		lineage := catalog.Lineage(document.Path)
		if len(lineage) != 3 {
			t.Fatalf("terminal lineage depth = %d, want 3", len(lineage))
		}
		for level := 0; level < len(lineage)-1; level++ {
			if lineage[level].ReplacementPath() != lineage[level+1].OldPath() ||
				lineage[level].ReplacementSHA256() != lineage[level+1].OldSHA256() {
				t.Fatalf("terminal lineage link %d differs", level)
			}
		}
	}
}

func TestDecodeRejectsIncompleteOrDowngradedV3Chain(t *testing.T) {
	encoded, _ := v3ChainFixture(t)
	var decoded catalogDocument
	decodeCatalogDocument(t, encoded, &decoded)
	decoded.Entries = decoded.Entries[:len(decoded.Entries)-1]
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("incomplete v3 chain was accepted")
	}

	decodeCatalogDocument(t, encoded, &decoded)
	decoded.ContractVersion = ContractVersionV2
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
		t.Fatal("v2 catalog accepted a third-level replacement")
	}
}
