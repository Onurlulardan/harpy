//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func claimPipelineEvent(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	authority OutboxAuthority,
	eventID uuid.UUID,
	wantGeneration int64,
) PublishClaim {
	t.Helper()
	claims, err := fixture.store.ClaimOutbox(
		ctx, authority, uuid.New(), 1, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].Envelope().EventID != eventID.String() ||
		claims[0].LeaseGeneration() != wantGeneration {
		t.Fatalf("pipeline claim = %#v, want event %s generation %d: %v",
			claims, eventID, wantGeneration, err)
	}
	return claims[0]
}

func installPipelineConfirmCrash(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	eventID uuid.UUID,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.pipeline_confirm_crash_target (
    event_id uuid PRIMARY KEY
);
CREATE FUNCTION platform.pipeline_fail_confirm_commit_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF NEW.delivery_state OPERATOR(pg_catalog.=) 'delivered' AND EXISTS (
        SELECT 1 FROM platform.pipeline_confirm_crash_target AS target
        WHERE target.event_id OPERATOR(pg_catalog.=) NEW.event_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5698',
            MESSAGE = 'injected outbox confirm commit crash';
    END IF;
    RETURN NEW;
END;
$function$;
CREATE CONSTRAINT TRIGGER pipeline_confirm_commit_crash
AFTER UPDATE ON platform.outbox
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION platform.pipeline_fail_confirm_commit_v1();
REVOKE ALL ON FUNCTION platform.pipeline_fail_confirm_commit_v1() FROM PUBLIC;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install confirm commit crash: %v", err)
	}
	if _, err := database.ExecContext(
		ctx, "INSERT INTO platform.pipeline_confirm_crash_target VALUES ($1)", eventID,
	); err != nil {
		t.Fatalf("arm confirm commit crash: %v", err)
	}
}

func removePipelineConfirmCrash(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
DROP TRIGGER pipeline_confirm_commit_crash ON platform.outbox;
DROP FUNCTION platform.pipeline_fail_confirm_commit_v1();
DROP TABLE platform.pipeline_confirm_crash_target;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("remove confirm commit crash: %v", err)
	}
}

func pipelineCommitCrash(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "P5698"
}

func assertPipelineOutboxInFlight(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	eventID uuid.UUID,
	wantGeneration int64,
) {
	t.Helper()
	var state string
	var generation int64
	var unconfirmed, leased bool
	err := database.QueryRowContext(ctx, `
SELECT delivery_state,lease_generation,
       confirm_sha256 IS NULL AND delivered_at IS NULL,
       lease_owner_id IS NOT NULL AND lease_until IS NOT NULL
FROM platform.outbox WHERE event_id=$1`, eventID).Scan(
		&state, &generation, &unconfirmed, &leased,
	)
	if err != nil || state != "publishing" || generation != wantGeneration || !unconfirmed || !leased {
		t.Fatalf("rolled back outbox = %s/%d unconfirmed=%t leased=%t: %v",
			state, generation, unconfirmed, leased, err)
	}
}

func assertPipelineFinalState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	eventID uuid.UUID,
	wantConfirm string,
) {
	t.Helper()
	var state, confirm string
	var attempts, effectCount, receiptCount, watermarkCount int
	err := database.QueryRowContext(ctx, `
SELECT delivery_state,delivery_attempts,confirm_sha256,
 (SELECT count(*) FROM pipeline_tenant_a.pipeline_effect_probe WHERE event_id=$1),
 (SELECT count(*) FROM platform.inbox
   WHERE record_kind='event_receipt' AND event_id=$1),
 (SELECT count(*) FROM platform.inbox
   WHERE record_kind='watermark' AND causation_id=$1)
FROM platform.outbox WHERE event_id=$1`, eventID).Scan(
		&state, &attempts, &confirm, &effectCount, &receiptCount, &watermarkCount,
	)
	if err != nil || state != "delivered" || attempts != 2 || confirm != wantConfirm ||
		effectCount != 1 || receiptCount != 1 || watermarkCount != 1 {
		t.Fatalf("final pipeline state=%s attempts=%d confirm=%t effect/inbox/watermark=%d/%d/%d: %v",
			state, attempts, confirm == wantConfirm, effectCount, receiptCount, watermarkCount, err)
	}
}
