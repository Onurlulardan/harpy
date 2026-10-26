package migration

import (
	"bytes"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBindDocumentsReturnsOnlySignedOrigin(t *testing.T) {
	manifest, selected, documents := testMigrationFixture(t)
	bound, err := bindDocuments(
		manifest, postgresv1.DatabaseOriginFresh, selected, documents,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(bound) != 1 || !bytes.Equal(bound[testFreshPath], testFreshSQL) {
		t.Fatal("bound migration source differs from signed fresh variant")
	}
	if _, leaked := bound[testAdoptedPath]; leaked {
		t.Fatal("non-selected database origin leaked into migration filesystem")
	}
	documents[0].content[0] = 'X'
	if bytes.Equal(bound[testFreshPath], documents[0].content) {
		t.Fatal("bound migration source aliases input document bytes")
	}
}

func TestBindDocumentsRejectsSetAndIdentityDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*postgresv1.MigrationManifest, *[]postgresv1.SelectedMigration, *[]Document)
	}{
		{"missing", func(_ *postgresv1.MigrationManifest, _ *[]postgresv1.SelectedMigration, docs *[]Document) {
			*docs = (*docs)[:1]
		}},
		{"extra", func(_ *postgresv1.MigrationManifest, _ *[]postgresv1.SelectedMigration, docs *[]Document) {
			*docs = append(*docs, mustDocument(t, testOwner,
				"core/internal/platform/migrator/migrations/fresh/20260713010100_extra.fresh.sql",
				testFreshSQL))
		}},
		{"duplicate", func(_ *postgresv1.MigrationManifest, _ *[]postgresv1.SelectedMigration, docs *[]Document) {
			*docs = append(*docs, (*docs)[0])
		}},
		{"owner", func(_ *postgresv1.MigrationManifest, _ *[]postgresv1.SelectedMigration, docs *[]Document) {
			(*docs)[0] = mustDocument(t, "platform_cutover", testFreshPath, testFreshSQL)
		}},
		{"scope", func(_ *postgresv1.MigrationManifest, _ *[]postgresv1.SelectedMigration, docs *[]Document) {
			(*docs)[0].scope = "tenant"
		}},
		{"selected origin", func(_ *postgresv1.MigrationManifest, selected *[]postgresv1.SelectedMigration, _ *[]Document) {
			(*selected)[0].Origin = postgresv1.DatabaseOriginV1Adopted
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest, selected, documents := testMigrationFixture(t)
			test.mutate(&manifest, &selected, &documents)
			if _, err := bindDocuments(
				manifest, postgresv1.DatabaseOriginFresh, selected, documents,
			); err == nil {
				t.Fatal("migration source drift was accepted")
			}
		})
	}
}
