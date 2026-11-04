//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"testing"
)

const iamSystemTransferAsyncAuthorityFixture = `
ALTER TABLE platform.domain_cutovers
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2',
    ADD COLUMN acceptance_digest text,
    ADD COLUMN changed_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp();
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';
ALTER ROLE harpy_v2_runtime LOGIN NOINHERIT
    PASSWORD 'runtime_integration_only';
CREATE ROLE harpy_cluster_admin NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE;`

func applyIAMSystemTransferAsyncPlatform(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, iamSystemTransferAsyncAuthorityFixture); err != nil {
		t.Fatalf("prepare IAM transfer async authority: %v", err)
	}
	documents, err := loadIAMSystemTransferAsyncDocuments(t, origin)
	if err != nil {
		t.Fatalf("load async migrations: %v", err)
	}
	sort.Slice(documents, func(left, right int) bool {
		return iamMigrationVersion(t, documents[left].path) <
			iamMigrationVersion(t, documents[right].path)
	})
	assertIAMAsyncRecoveryDocuments(t, documents)
	applyIAMPlatformDocuments(t, ctx, database, documents)
}

func loadIAMSystemTransferAsyncDocuments(
	t *testing.T,
	_ string,
) ([]iamFoundationDocument, error) {
	t.Helper()
	return nil, errors.New("async migration corpus is not a Goose source")
}

func assertIAMAsyncRecoveryDocuments(
	t *testing.T,
	documents []iamFoundationDocument,
) {
	t.Helper()
	required := map[string]bool{
		"20260715020100_": false,
		"20260718020139_": false,
		"20260718020140_": false,
		"20260718020143_": false,
	}
	for _, document := range documents {
		for marker := range required {
			if strings.Contains(document.path, marker) {
				required[marker] = true
			}
		}
	}
	for marker, found := range required {
		if !found {
			t.Fatalf("IAM async recovery migration %s is missing", marker)
		}
	}
}
