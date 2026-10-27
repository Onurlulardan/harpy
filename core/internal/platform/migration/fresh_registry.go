package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/gowebpki/jcs"
)

func newFreshRegistry(selected []Document) (*ReleaseRegistry, error) {
	manifest, err := freshManifest(selected)
	if err != nil {
		return nil, err
	}
	canonical, err := canonicalFreshManifest(manifest)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(canonical)
	paths := make([]string, len(selected))
	for index, document := range selected {
		paths[index] = document.Path()
	}
	registry := &ReleaseRegistry{
		manifests: map[string][]byte{
			hex.EncodeToString(digest[:]): canonical,
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

// The startup digest is the SHA-256 of the JCS canonical manifest.
func canonicalFreshManifest(manifest postgresv1.MigrationManifest) ([]byte, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("encode fresh migration manifest: %w", err)
	}
	if len(encoded) == 0 || len(encoded) > postgresv1.MaximumMigrationManifestBytes ||
		!utf8.Valid(encoded) {
		return nil, fmt.Errorf("fresh migration manifest size is invalid")
	}
	canonical, err := jcs.Transform(encoded)
	if err != nil {
		return nil, fmt.Errorf("canonicalize fresh migration manifest: %w", err)
	}
	if len(canonical) > postgresv1.MaximumMigrationManifestBytes {
		return nil, fmt.Errorf("fresh migration manifest size is invalid")
	}
	return canonical, nil
}

func freshManifest(selected []Document) (postgresv1.MigrationManifest, error) {
	if len(selected) == 0 || len(selected) > 10000 {
		return postgresv1.MigrationManifest{}, fmt.Errorf("fresh migration entries are invalid")
	}
	entries := make([]postgresv1.MigrationEntry, 0, len(selected))
	seenVersion := make(map[int64]struct{}, len(selected))
	seenBase := make(map[string]struct{}, len(selected))
	var previous int64
	for index, document := range selected {
		entry, err := freshEntry(document)
		if err != nil {
			return postgresv1.MigrationManifest{}, err
		}
		if _, exists := seenVersion[entry.Version]; exists {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"fresh migration version %d is duplicated", entry.Version,
			)
		}
		if index > 0 && entry.Version <= previous {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"fresh migration entries are not strictly version ordered",
			)
		}
		baseKey := document.Scope() + "/" + path.Base(document.Path())
		if _, exists := seenBase[baseKey]; exists {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"fresh migration basename %s is duplicated", path.Base(document.Path()),
			)
		}
		seenVersion[entry.Version] = struct{}{}
		seenBase[baseKey] = struct{}{}
		previous = entry.Version
		entries = append(entries, entry)
	}
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1",
		Engine:          "goose.v3",
		PostgresMajor:   18,
		HistoryTable:    "platform.platform_migration_history",
		Entries:         entries,
	}
	if err := manifest.Validate(); err != nil {
		return postgresv1.MigrationManifest{}, err
	}
	return manifest, nil
}

func freshEntry(document Document) (postgresv1.MigrationEntry, error) {
	metadata, err := InspectDocument(document)
	if err != nil {
		return postgresv1.MigrationEntry{}, fmt.Errorf(
			"inspect fresh migration %s: %w", document.Path(), err,
		)
	}
	version, err := freshVersion(document.Path())
	if err != nil {
		return postgresv1.MigrationEntry{}, err
	}
	if !strings.HasSuffix(path.Base(document.Path()), ".fresh.sql") {
		return postgresv1.MigrationEntry{}, fmt.Errorf(
			"fresh migration %s is not a fresh source", document.Path(),
		)
	}
	return postgresv1.MigrationEntry{
		Owner: document.Owner(), Scope: document.Scope(), Version: version,
		TransactionMode: metadata.TransactionMode, Phase: "expand",
		TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
		Variants: postgresv1.MigrationVariants{
			Fresh: &postgresv1.MigrationVariant{
				Path: document.Path(), SHA256: metadata.SHA256,
			},
		},
	}, nil
}
