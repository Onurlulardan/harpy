//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const layoutReplayRevision = "abababab-abab-7bab-8bab-abababababab"

func assertLayoutWriteIdempotency(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	definition := adapterDefinition
	created := createLayoutRevision(t, ctx, database, "layout_a",
		layoutReplayRevision, layoutapi.ResourceUser, definition)
	replayed := createLayoutRevision(t, ctx, database, "layout_a",
		layoutReplayRevision, layoutapi.ResourceUser, definition)
	if replayed.ID != created.ID || replayed.Version != created.Version ||
		replayed.RowRevision != created.RowRevision {
		t.Fatalf("create replay changed revision = %+v / %+v", created, replayed)
	}
	_, err := runLayoutRuntimeText(ctx, database, "layout_a", createSQL,
		layoutReplayRevision, string(layoutapi.ResourceRole), definition,
		layoutDigest(definition), layoutNow(42))
	assertLayoutSQLState(t, err, "40001")
	published := publishLayoutRevision(t, ctx, database, "layout_a",
		layoutReplayRevision, 1, definition, layoutNow(43))
	replayedPublication := publishLayoutRevision(t, ctx, database, "layout_a",
		layoutReplayRevision, 1, definition, layoutNow(44))
	if replayedPublication.Revision.RowRevision != 2 ||
		replayedPublication.Generation != published.Generation {
		t.Fatalf("publication replay changed state = %+v", replayedPublication)
	}
	drifted := layoutPublishArguments(
		layoutReplayRevision, 1, definition, layoutNow(44),
	)
	driftedCatalog := strings.Replace(
		layoutCatalog, `"mode":"editable"`, `"mode":"read_only"`, 1,
	)
	drifted[5], drifted[6] = driftedCatalog, layoutDigest(driftedCatalog)
	_, err = runLayoutRuntimeText(ctx, database, "layout_a", publishSQL, drifted...)
	assertLayoutSQLState(t, err, "40001")
}
