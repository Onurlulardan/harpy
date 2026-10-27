package main

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/gowebpki/jcs"
)

func schemaMigrationManifest(root string) ([]byte, error) {
	documents, err := schemaMigrationSources(root)
	if err != nil {
		return nil, err
	}
	selected, err := selectSchemaFresh(documents)
	if err != nil {
		return nil, err
	}
	return canonicalSchemaMigration(selected)
}

func selectSchemaFresh(documents []schemaSQL) ([]schemaSQL, error) {
	type key struct {
		owner, scope string
		version      int64
	}
	type candidate struct {
		document schemaSQL
		version  int64
		rank     int
	}
	best := map[key]candidate{}
	seen := map[string]struct{}{}
	for _, document := range documents {
		if strings.Contains(document.path, "/v1_adopted/") || strings.HasSuffix(document.path, ".v1_adopted.sql") {
			return nil, fmt.Errorf("v1_adopted is not present: %s", document.path)
		}
		version, err := schemaFreshVersion(document.path)
		if err != nil {
			return nil, err
		}
		rank, err := schemaSupersessionRank(document.path)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[document.path]; duplicate {
			return nil, fmt.Errorf("fresh migration %s is duplicated", document.path)
		}
		seen[document.path] = struct{}{}
		identity := key{document.owner, document.scope, version}
		current, exists := best[identity]
		if !exists || rank > current.rank {
			best[identity] = candidate{document, version, rank}
			continue
		}
		if rank == current.rank {
			return nil, fmt.Errorf("fresh migration version %d has ambiguous supersession", version)
		}
	}
	selected := make([]candidate, 0, len(best))
	for _, item := range best {
		selected = append(selected, item)
	}
	sort.Slice(selected, func(left, right int) bool {
		if selected[left].version != selected[right].version {
			return selected[left].version < selected[right].version
		}
		return selected[left].document.path < selected[right].document.path
	})
	documents = make([]schemaSQL, len(selected))
	for index, item := range selected {
		documents[index] = item.document
	}
	return documents, nil
}

func schemaFreshVersion(documentPath string) (int64, error) {
	base := path.Base(documentPath)
	prefix, _, ok := strings.Cut(base, "_")
	if !ok || len(prefix) != 14 {
		return 0, fmt.Errorf("fresh migration version is invalid: %s", documentPath)
	}
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("fresh migration version is invalid: %s", documentPath)
	}
	return version, nil
}

func schemaSupersessionRank(documentPath string) (int, error) {
	const marker = "/supersessions/"
	index := strings.Index(documentPath, marker)
	if index < 0 {
		return 0, nil
	}
	rest := documentPath[index+len(marker):]
	level, _, ok := strings.Cut(rest, "/")
	if !ok {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	if !strings.HasPrefix(level, "v") {
		return 1, nil
	}
	rank, err := strconv.Atoi(level[1:])
	if err != nil || rank < 1 {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	return rank, nil
}

func canonicalSchemaMigration(selected []schemaSQL) ([]byte, error) {
	entries := make([]postgresv1.MigrationEntry, 0, len(selected))
	seenVersion := map[int64]struct{}{}
	seenBase := map[string]struct{}{}
	var previous int64
	for index, document := range selected {
		metadata, err := postgresv1.InspectMigrationSource(document.body)
		if err != nil {
			return nil, fmt.Errorf("inspect fresh migration %s: %w", document.path, err)
		}
		version, err := schemaFreshVersion(document.path)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(path.Base(document.path), ".fresh.sql") {
			return nil, fmt.Errorf("fresh migration %s is not a fresh source", document.path)
		}
		if _, exists := seenVersion[version]; exists {
			return nil, fmt.Errorf("fresh migration version %d is duplicated", version)
		}
		if index > 0 && version <= previous {
			return nil, fmt.Errorf("fresh migration entries are not strictly version ordered")
		}
		baseKey := document.scope + "/" + path.Base(document.path)
		if _, exists := seenBase[baseKey]; exists {
			return nil, fmt.Errorf("fresh migration basename %s is duplicated", path.Base(document.path))
		}
		seenVersion[version] = struct{}{}
		seenBase[baseKey] = struct{}{}
		previous = version
		entries = append(entries, postgresv1.MigrationEntry{
			Owner: document.owner, Scope: document.scope, Version: version,
			TransactionMode: metadata.TransactionMode, Phase: "expand",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{Fresh: &postgresv1.MigrationVariant{
				Path: document.path, SHA256: metadata.SHA256,
			}},
		})
	}
	encoded, err := json.Marshal(postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.platform_migration_history",
		Entries: entries,
	})
	if err != nil {
		return nil, err
	}
	return jcs.Transform(encoded)
}
