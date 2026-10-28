package migration

import (
	_ "embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const releaseManifestSuffix = ".migration-manifest.json"

//go:embed release-index.json
var releaseIndexJSON []byte

func LoadReleaseRegistry() (*ReleaseRegistry, error) {
	archive, err := openDefaultReleaseArchive()
	if err != nil {
		return nil, err
	}
	return loadIndexedRelease(releaseIndexJSON, archive)
}

func loadIndexedRelease(indexDocument []byte, archive releaseReader) (*ReleaseRegistry, error) {
	index, known, contracts, err := parseReleaseIndex(indexDocument)
	if err != nil {
		return nil, err
	}
	if err = inventoryMatches(index.Digests, archive); err != nil {
		return nil, err
	}
	return &ReleaseRegistry{
		manifests:         map[string][]byte{},
		known:             known,
		digestList:        append([]string(nil), index.Digests...),
		documentContracts: contracts,
		readManifest:      archive.read,
	}, nil
}

func inventoryMatches(digests []string, archive releaseReader) error {
	names, err := archive.names()
	if err != nil {
		return err
	}
	found := make(map[string]struct{}, len(names))
	for _, name := range names {
		if !strings.HasSuffix(name, releaseManifestSuffix) || strings.Contains(name, "/") {
			return fmt.Errorf("migration release archive name is invalid")
		}
		digest := strings.TrimSuffix(name, releaseManifestSuffix)
		if _, duplicate := found[digest]; duplicate {
			return fmt.Errorf("migration release archive digest is duplicated")
		}
		found[digest] = struct{}{}
	}
	if len(found) != len(digests) {
		return fmt.Errorf("migration release archive does not match the release index")
	}
	for _, digest := range digests {
		if _, exists := found[digest]; !exists {
			return fmt.Errorf("migration release archive does not match the release index")
		}
	}
	return nil
}

func LoadReleaseManifestFS(files fs.FS, directory string) (*ReleaseRegistry, error) {
	entries, err := fs.ReadDir(files, directory)
	if err != nil {
		return nil, fmt.Errorf("read embedded migration manifests: %w", err)
	}
	registry := &ReleaseRegistry{
		manifests:         make(map[string][]byte, len(entries)),
		documentContracts: make(map[string]map[expectedDocument]struct{}),
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), releaseManifestSuffix) {
			return nil, fmt.Errorf("embedded migration manifest name is invalid")
		}
		content, err := fs.ReadFile(files, path.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		artifact, err := attestation.VerifyMigrationManifestArtifact(content)
		if err != nil {
			return nil, fmt.Errorf("verify embedded migration manifest: %w", err)
		}
		digest := strings.TrimSuffix(entry.Name(), releaseManifestSuffix)
		if digest != artifact.DigestSHA256() {
			return nil, fmt.Errorf("embedded migration manifest filename differs from digest")
		}
		manifest := artifact.Value()
		for _, migration := range manifest.Entries {
			if migration.Variants.Fresh == nil || migration.Variants.V1Adopted == nil {
				return nil, fmt.Errorf("embedded migration manifest has an incomplete origin pair")
			}
		}
		if _, duplicate := registry.manifests[digest]; duplicate {
			return nil, fmt.Errorf("embedded migration manifest digest is duplicated")
		}
		registry.manifests[digest] = artifact.CanonicalJSON()
		for documentPath, contract := range expectedDocuments(manifest) {
			contracts := registry.documentContracts[documentPath]
			if contracts == nil {
				contracts = make(map[expectedDocument]struct{})
				registry.documentContracts[documentPath] = contracts
			}
			contracts[contract] = struct{}{}
		}
	}
	if len(registry.manifests) == 0 {
		return nil, fmt.Errorf("embedded migration manifest registry is empty")
	}
	return registry, nil
}
