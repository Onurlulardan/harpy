//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncadoption "github.com/Onurlulardan/harpy/core/internal/platform/async/adoption"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const adoptionAuthorityFixture = `
CREATE SCHEMA platform;
CREATE TABLE platform.tenants (id uuid PRIMARY KEY);
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    state text NOT NULL,
    writer text NOT NULL,
    generation bigint NOT NULL,
    acceptance_digest text,
    changed_at timestamptz NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, domain)
);
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    state text NOT NULL,
    writer text NOT NULL,
    generation bigint NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id)
);
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id, domain)
);
INSERT INTO platform.tenants VALUES ('11111111-1111-4111-8111-111111111111');
INSERT INTO platform.cutover_groups VALUES (
    'tenant','11111111-1111-4111-8111-111111111111',
    'operations','quiescing','none',1
);
INSERT INTO platform.domain_cutovers VALUES (
    'tenant','11111111-1111-4111-8111-111111111111',
    'workflow','quiescing','none',1,NULL,pg_catalog.clock_timestamp()
);
INSERT INTO platform.cutover_group_members VALUES (
    'tenant','11111111-1111-4111-8111-111111111111','operations','workflow'
);`

func applyAdoptionMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, adoptionAuthorityFixture); err != nil {
		t.Fatalf("prepare adoption authority: %v", err)
	}
	asyncDocuments, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	cutoverDocuments, err := platformcutover.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := selectAdoptionDocuments(t, origin, asyncDocuments, cutoverDocuments)
	for _, document := range selected {
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func selectAdoptionDocuments(
	t *testing.T, origin string, asyncDocuments, cutoverDocuments []migration.Document,
) []migration.Document {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("unsupported origin %q", origin)
	}
	marker := "/migrations/" + origin + "/"
	selected := make([]migration.Document, 0, len(asyncDocuments)+1)
	for _, document := range asyncDocuments {
		version := adoptionDocumentVersion(t, document.Path())
		if strings.Contains(document.Path(), marker) &&
			version >= asyncadoption.AsyncMigrationFirstVersion &&
			version <= asyncadoption.AsyncMigrationLastVersion {
			selected = append(selected, document)
		}
	}
	for _, document := range cutoverDocuments {
		if strings.Contains(document.Path(), marker) &&
			adoptionDocumentVersion(t, document.Path()) == asyncadoption.CutoverFenceVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return adoptionDocumentVersion(t, selected[left].Path()) <
			adoptionDocumentVersion(t, selected[right].Path())
	})
	required := map[int64]bool{
		20260715021510: false, 20260715021520: false, 20260715021530: false,
		20260715021540: false, 20260715021542: false, 20260715021544: false,
		20260715021550: false, 20260715021552: false, 20260715021555: false,
		20260715021556: false, 20260715021557: false, 20260715021558: false,
		20260715021559: false, 20260715021605: false,
	}
	for _, document := range selected {
		if _, exists := required[adoptionDocumentVersion(t, document.Path())]; exists {
			required[adoptionDocumentVersion(t, document.Path())] = true
		}
	}
	for version, found := range required {
		if !found {
			t.Fatalf("adoption migration %d is missing for %s", version, origin)
		}
	}
	return selected
}

func adoptionDocumentVersion(t *testing.T, path string) int64 {
	t.Helper()
	base := path[strings.LastIndex(path, "/")+1:]
	version, err := strconv.ParseInt(base[:14], 10, 64)
	if err != nil {
		t.Fatalf("parse migration version %s: %v", path, err)
	}
	return version
}
