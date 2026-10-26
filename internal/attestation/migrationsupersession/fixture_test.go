package migrationsupersession

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const fixtureDirectory = "core/internal/platform/example/migrations"

func fixture(t *testing.T) ([]byte, []Source) {
	t.Helper()
	oldContent := []byte("-- +goose Up\nSELECT 1;\n")
	replacementContent := []byte("-- +goose Up\nSELECT 1::bigint;\n")
	oldHash := fmt.Sprintf("%x", sha256.Sum256(oldContent))
	replacementHash := fmt.Sprintf("%x", sha256.Sum256(replacementContent))
	document := catalogDocument{
		ContractVersion: ContractVersion,
		Owner:           "platform_example",
		Scope:           "platform",
		SourceDirectory: fixtureDirectory,
	}
	var sources []Source
	for _, origin := range []string{"fresh", "v1_adopted"} {
		base := "20260721010000_compile_pg18_v1." + origin + ".sql"
		oldPath := fixtureDirectory + "/" + origin + "/" + base
		replacementPath := fixtureDirectory + "/supersessions/" + origin + "/" + base
		document.Entries = append(document.Entries, entryDocument{
			OldPath: oldPath, OldSHA256: oldHash,
			ReplacementPath: replacementPath, ReplacementSHA256: replacementHash,
		})
		sources = append(sources,
			Source{Path: oldPath, Content: oldContent},
			Source{Path: replacementPath, Content: replacementContent},
		)
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return append(canonical, '\n'), sources
}

func decodeFixture(t *testing.T) (Catalog, []Source) {
	t.Helper()
	encoded, sources := fixture(t)
	catalog, err := DecodeCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, sources
}
