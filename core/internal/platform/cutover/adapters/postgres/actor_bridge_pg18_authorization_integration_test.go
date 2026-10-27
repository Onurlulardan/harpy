//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/jackc/pgx/v5"
)

func assertActorBridgeStaleGeneration(
	t *testing.T, ctx context.Context, v1 *pgx.Conn, current actorassertionv1.Claims,
) {
	t.Helper()
	staleDecision := current.Decision
	staleDecision.TargetGeneration--
	stale := signActorBridgeDecision(t, time.Now(), staleDecision)
	consumer := newActorBridgeTestConsumer(v1)
	if _, err := consumer.Consume(ctx, stale); !actorbridge.FaultIs(err, actorbridge.FaultAuthority) {
		t.Fatalf("stale actor bridge generation error = %v", err)
	}
}

func assertActorBridgeTenantIsolation(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	v1 *pgx.Conn,
	authorityA, authorityB actorbridge.Authority,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `
UPDATE t_actor_b.iam_user_permission_bindings
SET state='revoked', revision=revision+1, updated_at=clock_timestamp(),
    revoked_at=clock_timestamp()
WHERE id='97000000-0000-4000-8000-000000000001'`)
	if err != nil {
		t.Fatalf("revoke tenant B permission: %v", err)
	}
	consumer := newActorBridgeTestConsumer(v1)
	claimsB := issueActorBridgeClaims(t, time.Now(), actorBridgeTenantB, authorityB)
	v1B := connectActorBridgeV1(t, ctx, database.dsn, actorBridgeSchemaB)
	defer v1B.Close(ctx)
	if _, err := newActorBridgeTestConsumer(v1B).Consume(ctx, claimsB); !actorbridge.FaultIs(err, actorbridge.FaultUnauthorized) {
		t.Fatalf("tenant B revoked assertion error = %v", err)
	}

	claimsA := issueActorBridgeClaims(t, time.Now(), actorBridgeTenantA, authorityA)
	if _, err := consumer.Consume(ctx, claimsA); err != nil {
		t.Fatalf("tenant A permission changed with tenant B: %v", err)
	}
	var tenantA, tenantB int64
	if err := database.admin.QueryRow(ctx, `
SELECT count(*) FILTER (WHERE tenant_id=$1), count(*) FILTER (WHERE tenant_id=$2)
FROM platform.iam_v1_actor_assertion_receipts`,
		actorBridgeTenantA, actorBridgeTenantB).Scan(&tenantA, &tenantB); err != nil {
		t.Fatal(err)
	}
	if tenantA != 2 || tenantB != 0 {
		t.Fatalf("tenant receipt counts = %d/%d, want 2/0", tenantA, tenantB)
	}
}
