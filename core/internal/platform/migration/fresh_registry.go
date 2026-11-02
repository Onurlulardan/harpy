package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/gowebpki/jcs"
)

func newFreshRegistry(selected []Document) (*ReleaseRegistry, error) {
	entries := make([]postgresv1.MigrationEntry, 0, len(selected))
	paths := make([]string, len(selected))
	for index, document := range selected {
		version, err := freshDocumentVersion(document.Path())
		if err != nil {
			return nil, err
		}
		metadata, err := postgresv1.InspectMigrationSource(document.Content())
		if err != nil {
			return nil, fmt.Errorf("inspect fresh migration %s: %w", document.Path(), err)
		}
		entries = append(entries, postgresv1.MigrationEntry{
			Owner: document.Owner(), Scope: document.Scope(), Version: version,
			TransactionMode: metadata.TransactionMode, Phase: "expand",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{
				Fresh: &postgresv1.MigrationVariant{
					Path: document.Path(), SHA256: metadata.SHA256,
				},
			},
		})
		paths[index] = document.Path()
	}
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.goose_db_version",
		Entries: entries,
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("encode fresh migration manifest: %w", err)
	}
	canonical, err := jcs.Transform(encoded)
	if err != nil {
		return nil, fmt.Errorf("canonicalize fresh migration manifest: %w", err)
	}
	sum := sha256.Sum256(canonical)
	registry := &ReleaseRegistry{
		manifests:         map[string][]byte{hex.EncodeToString(sum[:]): canonical},
		documentContracts: make(map[string]map[expectedDocument]struct{}, len(selected)),
		freshPaths:        paths,
	}
	for documentPath, contract := range expectedDocuments(manifest) {
		registry.documentContracts[documentPath] = map[expectedDocument]struct{}{contract: {}}
	}
	if len(registry.documentContracts) != len(selected) {
		return nil, fmt.Errorf("fresh migration manifest entry count differs")
	}
	return registry, nil
}
