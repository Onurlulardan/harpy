package migration

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const releaseManifestSuffix = ".migration-manifest.json"

//go:embed release/*.migration-manifest.json
var releaseManifestFiles embed.FS

func LoadReleaseRegistry() (*ReleaseRegistry, error) {
	entries, err := fs.ReadDir(releaseManifestFiles, "release")
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
		content, err := fs.ReadFile(releaseManifestFiles, path.Join("release", entry.Name()))
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
