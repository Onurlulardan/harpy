package migrationbundle

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

func CanonicalFresh(root string) ([]byte, error) {
	documents, err := Collect(root)
	if err != nil {
		return nil, err
	}
	selected, err := Select(documents)
	if err != nil {
		return nil, err
	}
	_, canonical, err := Build(selected)
	return canonical, err
}

func Build(selected []Source) (postgresv1.MigrationManifest, []byte, error) {
	manifest, err := manifestFrom(selected)
	if err != nil {
		return postgresv1.MigrationManifest{}, nil, err
	}
	canonical, err := canonicalize(manifest)
	if err != nil {
		return postgresv1.MigrationManifest{}, nil, err
	}
	return manifest, canonical, nil
}

func Digest(canonical []byte) string {
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}

func manifestFrom(selected []Source) (postgresv1.MigrationManifest, error) {
	if len(selected) == 0 || len(selected) > 10000 {
		return postgresv1.MigrationManifest{}, fmt.Errorf("fresh migration entries are invalid")
	}
	entries := make([]postgresv1.MigrationEntry, 0, len(selected))
	seenVersion := make(map[int64]struct{}, len(selected))
	seenBase := make(map[string]struct{}, len(selected))
	var previous int64
	for index, source := range selected {
		entry, err := freshEntry(source)
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
		baseKey := source.Scope + "/" + path.Base(source.Path)
		if _, exists := seenBase[baseKey]; exists {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"fresh migration basename %s is duplicated", path.Base(source.Path),
			)
		}
		seenVersion[entry.Version] = struct{}{}
		seenBase[baseKey] = struct{}{}
		previous = entry.Version
		entries = append(entries, entry)
	}
	manifest := postgresv1.MigrationManifest{
		ContractVersion: freshManifestContract,
		Engine:          freshManifestEngine,
		PostgresMajor:   freshManifestPostgresMajor,
		HistoryTable:    freshManifestHistory,
		Entries:         entries,
	}
	if err := manifest.Validate(); err != nil {
		return postgresv1.MigrationManifest{}, err
	}
	return manifest, nil
}

func freshEntry(source Source) (postgresv1.MigrationEntry, error) {
	metadata, err := postgresv1.InspectMigrationSource(source.Body)
	if err != nil {
		return postgresv1.MigrationEntry{}, fmt.Errorf(
			"inspect fresh migration %s: %w", source.Path, err,
		)
	}
	version, err := freshVersion(source.Path)
	if err != nil {
		return postgresv1.MigrationEntry{}, err
	}
	if !strings.HasSuffix(path.Base(source.Path), ".fresh.sql") {
		return postgresv1.MigrationEntry{}, fmt.Errorf(
			"fresh migration %s is not a fresh source", source.Path,
		)
	}
	return postgresv1.MigrationEntry{
		Owner: source.Owner, Scope: source.Scope, Version: version,
		TransactionMode: metadata.TransactionMode, Phase: FreshAuthoringPhase,
		TargetDatabaseGeneration:     FreshAuthoringDatabaseGeneration,
		NextMinimumRuntimeGeneration: FreshAuthoringRuntimeGeneration,
		Variants: postgresv1.MigrationVariants{
			Fresh: &postgresv1.MigrationVariant{
				Path: source.Path, SHA256: metadata.SHA256,
			},
		},
	}, nil
}

func canonicalize(manifest postgresv1.MigrationManifest) ([]byte, error) {
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
