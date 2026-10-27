package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const (
	testVersion   = int64(20260713010000)
	testOwner     = "platform_migrator"
	testFreshPath = "core/internal/platform/migrator/migrations/fresh/" +
		"20260713010000_create_foundation.fresh.sql"
	testAdoptedPath = "core/internal/platform/migrator/migrations/v1_adopted/" +
		"20260713010000_create_foundation.v1_adopted.sql"
)

var (
	testFreshSQL   = []byte("-- +goose Up\nCREATE SCHEMA platform;\n")
	testAdoptedSQL = []byte("-- +goose Up\nALTER SCHEMA public RENAME TO platform;\n")
)

func testMigrationFixture(t *testing.T) (
	postgresv1.MigrationManifest,
	[]postgresv1.SelectedMigration,
	[]Document,
) {
	t.Helper()
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1",
		Engine:          "goose.v3",
		PostgresMajor:   18,
		HistoryTable:    "platform.platform_migration_history",
		Entries: []postgresv1.MigrationEntry{{
			Owner: testOwner, Scope: "platform", Version: testVersion,
			TransactionMode: "transactional", Phase: "foundation",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{
				Fresh: &postgresv1.MigrationVariant{
					Path: testFreshPath, SHA256: testSQLDigest(testFreshSQL),
				},
				V1Adopted: &postgresv1.MigrationVariant{
					Path: testAdoptedPath, SHA256: testSQLDigest(testAdoptedSQL),
				},
			},
		}},
	}
	selected, err := manifest.Select(postgresv1.DatabaseOriginFresh)
	if err != nil {
		t.Fatal(err)
	}
	documents := []Document{
		mustDocument(t, testOwner, testFreshPath, testFreshSQL),
		mustDocument(t, testOwner, testAdoptedPath, testAdoptedSQL),
	}
	return manifest, selected, documents
}

func mustDocument(t *testing.T, owner, name string, content []byte) Document {
	t.Helper()
	document, err := NewDocument(owner, "platform", name, content)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func testSQLDigest(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
