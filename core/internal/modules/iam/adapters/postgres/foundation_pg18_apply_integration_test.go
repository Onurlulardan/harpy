//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	iamPlatformFoundationFirst int64 = 20260716010000
	iamPlatformFoundationLast  int64 = 20260716012010
	iamTenantFoundationFirst   int64 = 20260716011000
	iamTenantFoundationLast    int64 = 20260716012900
)

type iamFoundationTenant struct {
	id     string
	schema string
}

type iamFoundationDocument struct {
	path    string
	content []byte
}

var iamFoundationTenants = []iamFoundationTenant{
	{id: "10000000-0000-4000-8000-000000000001", schema: "t_iam_a"},
	{id: "20000000-0000-4000-8000-000000000002", schema: "t_iam_b"},
}

func applyIAMFoundation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	prepareIAMFoundationSchemas(t, ctx, database, origin)
	platform := loadIAMFoundationDocuments(t, origin, "platform")
	applyIAMPlatformDocuments(t, ctx, database, platform)
	tenant := loadIAMFoundationDocuments(t, origin, "tenant")
	for _, fixture := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, fixture.schema, tenant)
	}
}

func prepareIAMFoundationSchemas(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	origin string,
) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime NOLOGIN;
CREATE ROLE harpy_adopter LOGIN NOINHERIT;
GRANT harpy_runtime TO harpy_v2_runtime;
CREATE SCHEMA platform;
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY,
    schema_name name NOT NULL UNIQUE
);
CREATE SCHEMA t_iam_a;
CREATE SCHEMA t_iam_b;
INSERT INTO platform.tenants (id, schema_name) VALUES
    ('10000000-0000-4000-8000-000000000001', 't_iam_a'),
    ('20000000-0000-4000-8000-000000000002', 't_iam_b');`
	if _, err := db.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare IAM schemas: %v", err)
	}
	prepareIAMPlatformAuthPrerequisites(t, ctx, db)
}

func loadIAMFoundationDocuments(
	t *testing.T,
	origin string,
	scope string,
) []iamFoundationDocument {
	t.Helper()
	if origin != "fresh" {
		t.Fatalf("unsupported IAM migration origin: %s", origin)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve IAM test directory: %v", err)
	}
	directory := filepath.Join(
		workingDirectory, "..", "..", "migrations", scope, origin,
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read IAM %s %s migrations: %v", scope, origin, err)
	}
	selected := make([]iamFoundationDocument, 0, 23)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version := iamMigrationVersion(t, entry.Name())
		platformMatch := scope == "platform" && version >= iamPlatformFoundationFirst &&
			version <= iamPlatformFoundationLast
		tenantMatch := scope == "tenant" && version >= iamTenantFoundationFirst &&
			version <= iamTenantFoundationLast
		if platformMatch || tenantMatch {
			path := filepath.Join(directory, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read IAM migration %s: %v", path, err)
			}
			selected = append(selected, iamFoundationDocument{path: path, content: content})
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].path < selected[right].path
	})
	want := 3
	if scope == "tenant" {
		want = 23
	}
	if len(selected) != want {
		t.Fatalf("%s %s IAM foundation count = %d, want %d", origin, scope, len(selected), want)
	}
	return selected
}

func iamMigrationVersion(t *testing.T, path string) int64 {
	t.Helper()
	base := filepath.Base(path)
	if len(base) < 14 {
		t.Fatalf("IAM migration path has no version: %s", path)
	}
	version, err := strconv.ParseInt(base[:14], 10, 64)
	if err != nil {
		t.Fatalf("parse IAM migration version %s: %v", path, err)
	}
	return version
}
