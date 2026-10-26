//go:build integration

package integrationfixture

import (
	"context"
	"database/sql"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const catalogDemoAnalyticsAuthorityID = "019c3300-0000-7000-8000-000000000071"

func (environment *CatalogDemoEnvironment) applyCatalogDemoAnalytics(
	t *testing.T,
	ctx context.Context,
	documents []migration.Document,
	origin string,
) {
	t.Helper()
	environment.execFixture(t, ctx, "analytics_platform_roles.fixture")
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].Path() < documents[right].Path()
	})
	platformCount := 0
	registered := false
	for _, document := range documents {
		path := filepath.ToSlash(document.Path())
		if !strings.Contains(path, "/migrations/platform/"+origin+"/") {
			continue
		}
		applyCatalogDemoDocument(t, ctx, environment.owner, document, false)
		platformCount++
		if strings.Contains(path,
			"20260718161004_create_analytics_current_tenant_authority_resolver_v1") {
			environment.registerCatalogDemoAnalyticsAuthority(t, ctx)
			registered = true
		}
	}
	tenantCount := applyCatalogDemoDocuments(
		t, ctx, environment.owner, documents,
		"/migrations/tenant/"+origin+"/", true,
	)
	if platformCount < 25 || tenantCount < 120 || !registered {
		t.Fatalf(
			"Analytics migration counts platform=%d tenant=%d authority=%v",
			platformCount, tenantCount, registered,
		)
	}
}

func applyCatalogDemoDocument(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	document migration.Document,
	tenant bool,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err == nil && tenant {
		_, err = transaction.ExecContext(
			ctx, string(catalogDemoFixture("set_search_path.fixture")),
		)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, string(document.Content()))
	}
	if err == nil {
		err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	if err != nil {
		t.Fatalf("apply %s: %v", document.Path(), err)
	}
}

func (environment *CatalogDemoEnvironment) registerCatalogDemoAnalyticsAuthority(
	t *testing.T, ctx context.Context,
) {
	t.Helper()
	transaction, err := environment.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(
		ctx,
		string(catalogDemoFixture("analytics_authority_session.fixture")),
	); err != nil {
		t.Fatalf("prepare Analytics authority registration: %v", err)
	}
	var authorityID string
	err = transaction.QueryRowContext(
		ctx,
		string(catalogDemoFixture("analytics_authority_register.fixture")),
		catalogDemoAnalyticsAuthorityID,
		CatalogDemoTenantID,
		CatalogDemoOwnerGeneration,
	).Scan(&authorityID)
	if err != nil || authorityID != catalogDemoAnalyticsAuthorityID {
		t.Fatalf(
			"register Catalog demo Analytics authority=%s err=%v",
			authorityID, err,
		)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit Catalog demo Analytics authority: %v", err)
	}
}
