//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestV1AsyncAdoptionRehearsalOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			owner := startAdoptionPG18(t, ctx, origin)
			applyAdoptionMigrations(t, ctx, owner, origin)
			grantAdoptionRoles(t, ctx, owner)
			adopter := openAdoptionRole(
				t, ctx, owner, "harpy_adopter", "adopter_integration_only",
			)
			cluster := openAdoptionRole(
				t, ctx, owner, "harpy_cluster_admin", "cluster_integration_only",
			)
			evidence := buildAdoptionEvidence(t)
			beginAndSealInventory(t, ctx, adopter, evidence)
			evidence = refreshAcceptanceAtDatabaseClock(t, ctx, owner, evidence)
			assertV2SwitchCode(t, ctx, owner, evidence.acceptance.AcceptanceSHA256, "P5667")
			adoptInventoryItem(t, ctx, adopter, evidence)
			assertZeroBeforeFencesFails(t, ctx, cluster, evidence)
			recordFences(t, ctx, cluster, evidence.fences)
			recordZeroAndAccept(t, ctx, cluster, evidence)
			assertV2SwitchCode(t, ctx, owner, strings.Repeat("f", 64), "P5667")
			assertV2SwitchCode(t, ctx, owner, evidence.acceptance.AcceptanceSHA256, "")
			assertAdoptionOutcome(t, ctx, owner, evidence)
		})
	}
}

func assertZeroBeforeFencesFails(
	t *testing.T, ctx context.Context, cluster *pgxpool.Pool,
	evidence adoptionEvidence,
) {
	t.Helper()
	zero := evidence.zero
	_, err := cluster.Exec(ctx, `SELECT platform.platform_async_record_v1_zero_inventory_v1(
        $1,$2,$3,$4,$5,$6,$7,$8)`, zero.RunID, zero.InventorySHA256,
		zero.QueueCount, zero.TimerCount, zero.FutureCount, zero.RunningCount,
		zero.ObservedAt, zero.ReceiptSHA256,
	)
	assertPGCode(t, err, "P5657")
}

func assertV2SwitchCode(
	t *testing.T, ctx context.Context, owner *sql.DB, acceptanceSHA, code string,
) {
	t.Helper()
	_, err := owner.ExecContext(ctx, `
UPDATE platform.domain_cutovers
SET state='v2', writer='v2', generation=2,
    acceptance_digest=$1, changed_at=pg_catalog.clock_timestamp()
WHERE scope_kind='tenant' AND scope_id=$2 AND domain=$3`,
		acceptanceSHA, adoptionScope, adoptionDomain,
	)
	if code == "" {
		if err != nil {
			t.Fatalf("switch accepted V2 authority: %v", err)
		}
		return
	}
	assertPGCode(t, err, code)
}

func assertAdoptionOutcome(
	t *testing.T, ctx context.Context, owner *sql.DB, evidence adoptionEvidence,
) {
	t.Helper()
	var runState, acceptanceSHA string
	if err := owner.QueryRowContext(ctx, `
SELECT state, acceptance_sha256 FROM platform.async_v1_adoption_runs WHERE run_id=$1`,
		evidence.inventory.RunID,
	).Scan(&runState, &acceptanceSHA); err != nil {
		t.Fatal(err)
	}
	if runState != "accepted" || acceptanceSHA != evidence.acceptance.AcceptanceSHA256 {
		t.Fatalf("adoption outcome = %s/%s", runState, acceptanceSHA)
	}
	var writer, state string
	var generation int64
	if err := owner.QueryRowContext(ctx, `
SELECT writer_origin, state, cutover_generation FROM platform.jobs WHERE id=$1`,
		adoptionJobID,
	).Scan(&writer, &state, &generation); err != nil {
		t.Fatal(err)
	}
	if writer != "v1" || state != "queued" || generation != 2 {
		t.Fatalf("adopted job = %s/%s/%d", writer, state, generation)
	}
}

func assertPGCode(t *testing.T, err error, want string) {
	t.Helper()
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != want {
		t.Fatalf("PostgreSQL code = %v, want %s", err, want)
	}
}
