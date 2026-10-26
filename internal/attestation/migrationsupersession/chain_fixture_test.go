package migrationsupersession

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func chainFixture(t *testing.T) ([]byte, []Source) {
	t.Helper()
	contents := [][]byte{
		[]byte("-- +goose Up\nSELECT 1;\n"),
		[]byte("-- +goose Up\nSELECT 1::bigint;\n"),
		[]byte("-- +goose Up\nSELECT 1::numeric;\n"),
	}
	hashes := make([]string, len(contents))
	for index, content := range contents {
		hashes[index] = fmt.Sprintf("%x", sha256.Sum256(content))
	}
	document := catalogDocument{
		ContractVersion: ContractVersionV2,
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
		}
		for level := 0; level < 2; level++ {
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

func decodeChainFixture(t *testing.T) (Catalog, []Source) {
	t.Helper()
	encoded, sources := chainFixture(t)
	catalog, err := DecodeCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, sources
}
