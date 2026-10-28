package migrationmanifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func TestResolveSupersessionCatalogSupportsScopedAndLegacyOwners(t *testing.T) {
	source := "owner/migrations/platform"
	for _, candidate := range []struct {
		name     string
		location string
	}{
		{"scoped", filepath.Join(filepath.FromSlash(source),
			migrationsupersession.CatalogFileName)},
		{"legacy", filepath.Join("owner", "migrations",
			migrationsupersession.CatalogFileName)},
	} {
		t.Run(candidate.name, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, candidate.location)
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			path, info, found, err := resolveSupersessionCatalog(root, source)
			if err != nil || !found || info == nil ||
				filepath.Clean(filepath.FromSlash(path)) !=
					filepath.Clean(candidate.location) {
				t.Fatalf("resolved path=%q found=%t info=%v err=%v",
					path, found, info, err)
			}
		})
	}
}

func TestResolveSupersessionCatalogRejectsAmbiguousLocations(t *testing.T) {
	root := t.TempDir()
	source := "owner/migrations/platform"
	for _, location := range []string{
		filepath.Join(filepath.FromSlash(source),
			migrationsupersession.CatalogFileName),
		filepath.Join("owner", "migrations", migrationsupersession.CatalogFileName),
	} {
		target := filepath.Join(root, location)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, _, err := resolveSupersessionCatalog(root, source); err == nil {
		t.Fatal("ambiguous supersession catalogs were accepted")
	}
}
