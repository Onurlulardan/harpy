package migrationmanifest

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestValidateReleaseFrontierAdvanceRequiresGenerationBump(t *testing.T) {
	directory := t.TempDir()
	previous := releaseFrontierFixture(13)
	writeReleaseFrontier(t, directory, previous)

	current := appendReleaseFrontier(previous, 13)
	if err := ValidateReleaseFrontierAdvance(directory, current); err == nil ||
		!strings.Contains(err.Error(), "target_database_generation") {
		t.Fatalf("database generation reuse error = %v", err)
	}
	current.Entries[1].TargetDatabaseGeneration = 14
	if err := ValidateReleaseFrontierAdvance(directory, current); err != nil {
		t.Fatalf("advanced database generation was rejected: %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceRejectsInterposition(t *testing.T) {
	directory := t.TempDir()
	base := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, base)
	published := appendReleaseFrontierAt(
		base, 20260722000002, 13, "b",
	)
	writeReleaseFrontier(t, directory, published)

	current := appendReleaseFrontierAt(
		base, 20260722000001, 13, "c",
	)
	current = appendReleaseFrontierAt(
		current, 20260722000002, 13, "b",
	)
	current = appendReleaseFrontierAt(
		current, 20260722000003, 14, "d",
	)
	err := ValidateReleaseFrontierAdvance(directory, current)
	if err == nil || !strings.Contains(err.Error(), "maximal immutable release") {
		t.Fatalf("interposed migration lineage error = %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceRejectsAmbiguousMaximalBranch(t *testing.T) {
	directory := t.TempDir()
	base := releaseFrontierFixture(12)
	branchA := appendReleaseFrontierAt(
		base, 20260722000001, 13, "b",
	)
	branchB := appendReleaseFrontierAt(
		base, 20260722000001, 13, "c",
	)
	writeReleaseFrontier(t, directory, branchA)
	writeReleaseFrontier(t, directory, branchB)

	current := appendReleaseFrontierAt(
		branchA, 20260722000002, 14, "d",
	)
	err := ValidateReleaseFrontierAdvance(directory, current)
	if err == nil || !strings.Contains(err.Error(), "lineage is ambiguous") {
		t.Fatalf("ambiguous migration lineage error = %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceUsesOnlyMaximalHistoricalFrontier(t *testing.T) {
	directory := t.TempDir()
	staleBranch := releaseFrontierFixture(12)
	staleBranch.Entries[0].Variants.Fresh.SHA256 = strings.Repeat("e", 64)
	writeReleaseFrontier(t, directory, staleBranch)

	base := releaseFrontierFixture(12)
	maximal := appendReleaseFrontierAt(
		base, 20260722000001, 13, "b",
	)
	writeReleaseFrontier(t, directory, maximal)
	current := appendReleaseFrontierAt(
		maximal, 20260722000002, 14, "c",
	)
	if err := ValidateReleaseFrontierAdvance(directory, current); err != nil {
		t.Fatalf("maximal immutable frontier was not authoritative: %v", err)
	}
}

func writeReleaseFrontier(
	t *testing.T,
	directory string,
	manifest postgresv1.MigrationManifest,
) {
	t.Helper()
	canonical, digest, err := Encode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, digest+immutableManifestSuffix)
	if err := os.WriteFile(path, canonical, 0o600); err != nil {
		t.Fatal(err)
	}
}

func releaseFrontierFixture(generation int64) postgresv1.MigrationManifest {
	return postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.platform_migration_history",
		Entries: []postgresv1.MigrationEntry{
			releaseFrontierEntry(20260722000000, generation, "a"),
		},
	}
}

func appendReleaseFrontier(
	previous postgresv1.MigrationManifest,
	generation int64,
) postgresv1.MigrationManifest {
	version := previous.Entries[len(previous.Entries)-1].Version + 1
	return appendReleaseFrontierAt(previous, version, generation, "b")
}

func appendReleaseFrontierAt(
	previous postgresv1.MigrationManifest,
	version int64,
	generation int64,
	digestCharacter string,
) postgresv1.MigrationManifest {
	current := previous
	current.Entries = append(
		append([]postgresv1.MigrationEntry(nil), previous.Entries...),
		releaseFrontierEntry(version, generation, digestCharacter),
	)
	return current
}

func releaseFrontierEntry(
	version int64,
	generation int64,
	digestCharacter string,
) postgresv1.MigrationEntry {
	base := "core/internal/modules/forms/migrations/platform/"
	versionText := strconv.FormatInt(version, 10)
	digest := strings.Repeat(digestCharacter, 64)
	return postgresv1.MigrationEntry{
		Owner: "forms", Scope: "platform", Version: version,
		TransactionMode: "transactional", Phase: "expand",
		TargetDatabaseGeneration: generation, NextMinimumRuntimeGeneration: 13,
		Variants: postgresv1.MigrationVariants{
			Fresh: &postgresv1.MigrationVariant{
				Path: base + "fresh/" + versionText + "_guard.fresh.sql", SHA256: digest,
			},
			V1Adopted: &postgresv1.MigrationVariant{
				Path:   base + "v1_adopted/" + versionText + "_guard.v1_adopted.sql",
				SHA256: digest,
			},
		},
	}
}
