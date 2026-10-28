package migrationmanifest

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

type supersessionCatalogFixture struct {
	ContractVersion string                     `json:"contract_version"`
	Entries         []supersessionEntryFixture `json:"entries"`
	Owner           string                     `json:"owner"`
	Scope           string                     `json:"scope"`
	SourceDirectory string                     `json:"source_directory"`
}

type supersessionEntryFixture struct {
	OldPath           string `json:"old_path"`
	OldSHA256         string `json:"old_sha256"`
	ReplacementPath   string `json:"replacement_path"`
	ReplacementSHA256 string `json:"replacement_sha256"`
}

func TestCollectSelectsExactPairedSupersession(t *testing.T) {
	root := t.TempDir()
	directory := "core/internal/platform/example/migrations"
	writeSupersessionFixture(t, root, directory)
	group, err := Collect(root, SourceSpec{
		Owner: testOwner, Scope: "platform", Phase: "foundation",
		TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
		VersionFrom: 20260713010000, VersionTo: 20260713010000,
		Directory: directory,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(group.Documents) != 2 {
		t.Fatalf("selected documents = %d", len(group.Documents))
	}
	for _, document := range group.Documents {
		if !strings.Contains(document.Path(), "/supersessions/") ||
			!strings.Contains(string(document.Content()), "bigint") {
			t.Fatalf("old migration remained selected: %s", document.Path())
		}
	}
	if _, err := Build(testStorageInventory(group), []SourceGroup{group}); err != nil {
		t.Fatal(err)
	}
}

func writeSupersessionFixture(t *testing.T, root, directory string) {
	t.Helper()
	oldContent := []byte("-- +goose Up\nSELECT 1;\n")
	replacementContent := []byte("-- +goose Up\nSELECT 1::bigint;\n")
	catalog := supersessionCatalogFixture{
		ContractVersion: migrationsupersession.ContractVersion,
		Owner:           testOwner, Scope: "platform", SourceDirectory: directory,
	}
	for _, origin := range []string{"fresh", "v1_adopted"} {
		base := testVersion + "_create_foundation." + origin + ".sql"
		oldPath := directory + "/" + origin + "/" + base
		replacementPath := directory + "/supersessions/" + origin + "/" + base
		writeFixtureSource(t, root, oldPath, oldContent)
		writeFixtureSource(t, root, replacementPath, replacementContent)
		catalog.Entries = append(catalog.Entries, supersessionEntryFixture{
			OldPath: oldPath, OldSHA256: fixtureSHA256(oldContent),
			ReplacementPath:   replacementPath,
			ReplacementSHA256: fixtureSHA256(replacementContent),
		})
	}
	encoded, err := attestation.MarshalCanonical(catalog)
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	catalogPath := filepath.Join(root, filepath.FromSlash(filepath.Dir(directory)),
		migrationsupersession.CatalogFileName)
	if err := os.WriteFile(catalogPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeFixtureSource(t *testing.T, root, name string, content []byte) {
	t.Helper()
	absolute := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(absolute), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolute, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func fixtureSHA256(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
