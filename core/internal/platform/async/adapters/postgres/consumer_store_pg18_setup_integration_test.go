//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const consumerStoreLastVersion int64 = 20260715021500

func TestConsumerStoreSQLAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx := context.Background()
			database := startAsyncMigrationDatabase(t, ctx, "consumer_"+origin)
			prepareAsyncAuthorityFixture(t, ctx, database)
			addAsyncAuthorityStateColumns(t, ctx, database)
			applyConsumerStoreMigrations(t, ctx, database, origin)
			prepareConsumerStoreRuntime(t, ctx, database)
			assertConsumerStoreCases(t, ctx, database)
		})
	}
}

func applyConsumerStoreMigrations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]migration.Document, 0, 32)
	marker := "/migrations/" + origin + "/"
	for _, document := range documents {
		version := asyncDocumentVersion(t, document.Path())
		if strings.Contains(document.Path(), marker) &&
			version >= asyncSliceAFirstVersion && version <= consumerStoreLastVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	for _, document := range selected {
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func prepareConsumerStoreRuntime(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_v2_runtime;
GRANT USAGE ON SCHEMA platform TO harpy_v2_runtime;
GRANT SELECT ON TABLE platform.outbox, platform.inbox,
    platform.idempotency_receipts TO harpy_v2_runtime;
GRANT INSERT ON TABLE platform.outbox, platform.inbox,
    platform.idempotency_receipts TO harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_assert_current_authority_v1(
    text,uuid,text,bigint,text,bigint
) TO harpy_v2_runtime;
CREATE TABLE platform.consumer_effect_probe (
    event_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    effect_kind text NOT NULL
);
GRANT SELECT, INSERT ON TABLE platform.consumer_effect_probe
TO harpy_v2_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare consumer runtime: %v", err)
	}
}

func consumerRuntimeConnection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		connection.Close()
		t.Fatalf("activate consumer runtime: %v", err)
	}
	t.Cleanup(func() {
		_, _ = connection.ExecContext(context.Background(), "RESET SESSION AUTHORIZATION")
		_ = connection.Close()
	})
	return connection
}
