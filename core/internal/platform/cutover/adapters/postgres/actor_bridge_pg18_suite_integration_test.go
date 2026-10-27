//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/google/uuid"
)

func TestIAMV1ActorBridgeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			database := startCutoverIntegrationDatabase(t, ctx)
			prepareActorBridgeIAM(t, ctx, database, origin)
			assertActorBridgeLegacySentinels(t, ctx, database.admin, origin)

			v2 := connectContractRetirementRuntime(t, ctx, database.dsn)
			defer v2.Close(ctx)
			v1 := connectActorBridgeV1(t, ctx, database.dsn, actorBridgeSchemaA)
			defer v1.Close(ctx)
			assertActorBridgeSessionIdentity(t, ctx, v1, "harpy_v1_runtime")

			issuer, err := NewActorBridgeDatabase(v2)
			if err != nil {
				t.Fatal(err)
			}
			authorityA := resolveActorBridgeAuthority(
				t, ctx, issuer, actorBridgeTenantA,
			)
			issuedA := issueActorBridgeCredential(t, time.Now(), actorBridgeTenantA, authorityA)
			claimsA := issuedA.claims
			assertActorBridgeDirectDMLClosed(t, ctx, v1)
			assertActorBridgeRollbackCommitReplay(t, ctx, database, v1, issuedA)
			assertActorBridgeStaleGeneration(t, ctx, v1, claimsA)

			authorityB := resolveActorBridgeAuthority(
				t, ctx, issuer, actorBridgeTenantB,
			)
			assertActorBridgeTenantIsolation(
				t, ctx, database, v1, authorityA, authorityB,
			)
			assertActorBridgeConcurrentReplay(
				t, ctx, database.dsn, authorityA,
			)
			assertActorBridgeMaintenance(t, ctx, database, issuer, 3)
		})
	}
}

func resolveActorBridgeAuthority(
	t *testing.T,
	ctx context.Context,
	issuer *ActorBridgeDatabase,
	tenantID string,
) actorbridge.Authority {
	t.Helper()
	authority, err := issuer.ResolveIssueAuthority(
		ctx, uuid.MustParse(tenantID), actorassertionv1.DomainForms,
	)
	if err != nil {
		t.Fatal(err)
	}
	if authority.IAMGroupGeneration() != 9 || authority.IAMGeneration() != 7 ||
		authority.TargetGeneration() != 11 {
		t.Fatalf("actor bridge authority = %+v", authority)
	}
	return authority
}

func assertActorBridgeSessionIdentity(
	t *testing.T, ctx context.Context, connection ActorBridgeQueryer, wantSession string,
) {
	t.Helper()
	var sessionUser, currentUser string
	if err := connection.QueryRow(ctx,
		"SELECT session_user::text,current_user::text",
	).Scan(&sessionUser, &currentUser); err != nil {
		t.Fatal(err)
	}
	if sessionUser != wantSession || currentUser != "harpy_runtime" {
		t.Fatalf("actor bridge identity = %s/%s", sessionUser, currentUser)
	}
}
