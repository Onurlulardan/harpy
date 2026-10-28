package migrationmanifest

import (
	"strconv"
	"strings"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestEncodeUsesMigrationArtifactSizeLimit(t *testing.T) {
	entries := make([]postgresv1.MigrationEntry, 200)
	start := time.Date(2026, 7, 13, 1, 0, 0, 0, time.UTC)
	directory := "core/internal/platform/" + strings.Repeat("migration_", 30)
	for index := range entries {
		version := start.Add(time.Duration(index) * time.Second).Format("20060102150405")
		entries[index] = postgresv1.MigrationEntry{
			Owner: "platform_migrator", Scope: "platform",
			Version: mustMigrationVersion(t, version), TransactionMode: "transactional",
			Phase: "foundation", TargetDatabaseGeneration: 1,
			NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{
				Fresh: &postgresv1.MigrationVariant{
					Path:   directory + "/" + version + "_fixture.fresh.sql",
					SHA256: strings.Repeat("a", 64),
				},
				V1Adopted: &postgresv1.MigrationVariant{
					Path:   directory + "/" + version + "_fixture.v1_adopted.sql",
					SHA256: strings.Repeat("b", 64),
				},
			},
		}
	}
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.platform_migration_history",
		Entries: entries,
	}
	canonical, digest, err := Encode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical) <= runtimev1.MaximumAttestationDocumentBytes {
		t.Fatalf("fixture did not cross generic attestation bound: %d", len(canonical))
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(canonical)
	if err != nil || verified.DigestSHA256() != digest {
		t.Fatalf("large migration artifact did not verify: %v", err)
	}
}

func mustMigrationVersion(t *testing.T, value string) int64 {
	t.Helper()
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
