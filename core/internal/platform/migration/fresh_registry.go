package migration

import (
	"fmt"

	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func newFreshRegistry(selected []Document) (*ReleaseRegistry, error) {
	sources := make([]migrationbundle.Source, len(selected))
	for index, document := range selected {
		sources[index] = migrationbundle.Source{
			Owner: document.Owner(), Scope: document.Scope(),
			Path: document.Path(), Body: document.Content(),
		}
	}
	manifest, canonical, err := migrationbundle.Build(sources)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(selected))
	for index, document := range selected {
		paths[index] = document.Path()
	}
	registry := &ReleaseRegistry{
		manifests: map[string][]byte{
			migrationbundle.Digest(canonical): canonical,
		},
		documentContracts: make(map[string]map[expectedDocument]struct{}, len(selected)),
		freshPaths:        paths,
	}
	for documentPath, contract := range expectedDocuments(manifest) {
		registry.documentContracts[documentPath] = map[expectedDocument]struct{}{
			contract: {},
		}
	}
	if len(registry.documentContracts) != len(selected) {
		return nil, fmt.Errorf("fresh migration manifest entry count differs")
	}
	return registry, nil
}
