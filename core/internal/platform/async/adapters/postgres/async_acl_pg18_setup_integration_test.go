//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/jackc/pgx/v5/pgxpool"
)

const asyncACLRoleFixture = `
CREATE ROLE harpy_runtime NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE;
CREATE ROLE harpy_v2_runtime LOGIN PASSWORD 'runtime_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
GRANT harpy_runtime TO harpy_v2_runtime
    WITH INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_adopter LOGIN PASSWORD 'adopter_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE harpy_cluster_admin LOGIN PASSWORD 'cluster_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE harpy_migrator LOGIN PASSWORD 'migrator_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
GRANT harpy_owner TO harpy_migrator
    WITH INHERIT FALSE, SET TRUE;`

func TestAsyncRuntimeACLOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAdoptionPG18(t, ctx, "acl_"+origin)
			applyAsyncACLFixture(t, ctx, database, origin)
			disableAsyncOwnerLogin(t, ctx, database)
			runtime := openAsyncRuntimeRole(t, ctx, database)
			assertAsyncACLPrivileges(t, ctx, runtime, database)
			assertAsyncACLRetentionEntrypoints(t, ctx, runtime, database)
			assertAsyncACLJobLifecycle(t, ctx, runtime)
			assertAsyncACLDenials(t, ctx, runtime)
		})
	}
}

func applyAsyncACLFixture(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, asyncACLRoleFixture); err != nil {
		t.Fatalf("create async ACL roles: %v", err)
	}
	if _, err := database.ExecContext(ctx, adoptionAuthorityFixture); err != nil {
		t.Fatalf("create async ACL authority fixture: %v", err)
	}
	if _, err := database.ExecContext(ctx, `
UPDATE platform.cutover_groups SET state='v2', writer='v2';
UPDATE platform.domain_cutovers SET state='v2', writer='v2';`); err != nil {
		t.Fatalf("activate async ACL authority: %v", err)
	}
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatalf("load async ACL migrations: %v", err)
	}
	marker := "/migrations/" + origin + "/"
	selected := make([]int, 0, len(documents)/2)
	for index, document := range documents {
		if strings.Contains(document.Path(), marker) {
			selected = append(selected, index)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return documents[selected[left]].Path() < documents[selected[right]].Path()
	})
	required := map[string]bool{
		"20260715021600_": false,
		"20260715021630_": false,
		"20260715021720_": false,
	}
	for _, index := range selected {
		for marker := range required {
			if strings.Contains(documents[index].Path(), marker) {
				required[marker] = true
			}
		}
	}
	for marker, found := range required {
		if !found {
			t.Fatalf("%s async migration %s is missing", origin, marker)
		}
	}
	for _, index := range selected {
		document := documents[index]
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func openAsyncRuntimeRole(
	t *testing.T, ctx context.Context, database *sql.DB,
) *pgxpool.Conn {
	t.Helper()
	pool := openAdoptionRole(
		t, ctx, database, "harpy_v2_runtime", "runtime_integration_only",
	)
	connection, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire async runtime connection: %v", err)
	}
	t.Cleanup(connection.Release)
	if _, err := connection.Exec(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatalf("set async runtime role: %v", err)
	}
	return connection
}
