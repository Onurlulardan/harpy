//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertIAMOrganizationOwnerParity(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const organizationID = "b2000000-0000-4000-8000-000000000001"
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	profile, profileSHA := []byte(`{}`), sha256.Sum256([]byte(`{}`))
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		callIAMIdentityMutation(t, ctx, tx, iamIdentityOrganizationCreate,
			"b2100000-0000-4000-8000-000000000001", bytes.Repeat([]byte{0x81}, 32),
			organizationID, "Owned Organization", "owned-organization",
			iamIdentityOrgSchemaID, profile, profileSHA[:], iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
	})
	mustExecIAMTenant(t, ctx, database, "t_iam_a", iamOrganizationOwnerGraphFixture)
	request := bytes.Repeat([]byte{0x82}, 32)
	ownerIDs := []uuid.UUID{uuid.MustParse(iamUserTwoID)}
	membershipIDs := []uuid.UUID{
		uuid.MustParse("b2000000-0000-4000-8000-000000000005"),
	}
	callIAMOrganizationOwnerUpdate(t, ctx, database, actor, organizationID,
		"b2100000-0000-4000-8000-000000000002", request,
		ownerIDs, membershipIDs, false)
	first := readIAMOrganizationOwnerState(t, ctx, database, organizationID)
	callIAMOrganizationOwnerUpdate(t, ctx, database, actor, organizationID,
		"b2100000-0000-4000-8000-000000000002", request,
		ownerIDs, membershipIDs, true)
	second := readIAMOrganizationOwnerState(t, ctx, database, organizationID)
	if first != second || first !=
		"51000000-0000-4000-8000-000000000002:owner:active:b2000000-0000-4000-8000-000000000002:2,"+
			"51000000-0000-4000-8000-000000000003:member:active:b2000000-0000-4000-8000-000000000001:2" {
		t.Fatalf("organization owner replacement=%q replay=%q", first, second)
	}
	assertIAMOrganizationOwnerReceiptAndProtection(t, ctx, database, organizationID)
	assertIAMOrganizationOwnerFailures(t, ctx, database, actor, organizationID,
		profile, profileSHA[:])
}

func callIAMOrganizationOwnerUpdate(
	t *testing.T, ctx context.Context, database *sql.DB, actor iamIdentityActor,
	organizationID, operationID string, request []byte,
	ownerIDs, membershipIDs []uuid.UUID, wantReplay bool,
) {
	t.Helper()
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		var id string
		var revision, clock int64
		var replayed bool
		profileSHA := sha256.Sum256([]byte(`{}`))
		err := tx.QueryRowContext(ctx, iamOrganizationOwnerUpdateSQL,
			operationID, request, organizationID, int64(1),
			"Owned Organization V2", "owned-organization-v2", ownerIDs, membershipIDs,
			iamIdentityOrgSchemaID, []byte(`{}`), profileSHA[:],
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion,
		).Scan(&id, &revision, &clock, &replayed)
		if err != nil || id != organizationID || revision != 2 || replayed != wantReplay {
			t.Fatalf("owner update=%s/%d/%d/%t error=%v", id, revision, clock, replayed, err)
		}
	})
}

func readIAMOrganizationOwnerState(
	t *testing.T, ctx context.Context, database *sql.DB, organizationID string,
) string {
	t.Helper()
	connection, tx := beginIAMTenantTx(t, ctx, database, "t_iam_a")
	defer connection.Close()
	defer tx.Rollback()
	var state string
	err := tx.QueryRowContext(ctx, `SELECT pg_catalog.string_agg(
 user_id::text||':'||membership_kind||':'||state||':'||id::text||':'||revision,
 ',' ORDER BY user_id) FROM iam_organization_memberships WHERE organization_id=$1`,
		organizationID).Scan(&state)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func assertIAMOrganizationOwnerReceiptAndProtection(
	t *testing.T, ctx context.Context, database *sql.DB, organizationID string,
) {
	t.Helper()
	connection, tx := beginIAMTenantTx(t, ctx, database, "t_iam_a")
	defer connection.Close()
	defer tx.Rollback()
	var receipts, assignments int
	err := tx.QueryRowContext(ctx, `SELECT
 (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
  WHERE operation_id='b2100000-0000-4000-8000-000000000002'),
 (SELECT pg_catalog.count(*) FROM iam_user_role_assignments
  WHERE organization_id=$1 AND membership_id=$1 AND state='active')`,
		organizationID).Scan(&receipts, &assignments)
	if err != nil || receipts != 1 || assignments != 1 {
		t.Fatalf("owner receipt/protected assignment=%d/%d error=%v",
			receipts, assignments, err)
	}
}

func assertIAMOrganizationOwnerFailures(
	t *testing.T, ctx context.Context, database *sql.DB, actor iamIdentityActor,
	organizationID string, profile, profileSHA []byte,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, "t_iam_b", iamCrossTenantOwnerFixture)
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		row := tx.QueryRowContext(ctx, iamOrganizationNoOwnerCreateSQL,
			"b2200000-0000-4000-8000-000000000000", bytes.Repeat([]byte{0x83}, 32),
			"b2000000-0000-4000-8000-000000000010", "Ownerless", "ownerless",
			iamIdentityOrgSchemaID, profile, profileSHA, iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		var id string
		var revision, clock int64
		var replayed bool
		assertIAMIdentitySQLState(t, row.Scan(&id, &revision, &clock, &replayed), "22023")
	})
	tests := []struct {
		operation, state string
		marker           byte
		owners           []uuid.UUID
		memberships      []uuid.UUID
	}{
		{"b2200000-0000-4000-8000-000000000001", "22023", 0x84,
			[]uuid.UUID{}, []uuid.UUID{}},
		{"b2200000-0000-4000-8000-000000000002", "23514", 0x85,
			[]uuid.UUID{uuid.MustParse("51000000-0000-4000-8000-000000000099")},
			[]uuid.UUID{uuid.MustParse("b2000000-0000-4000-8000-000000000099")}},
	}
	for _, test := range tests {
		probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
			row := tx.QueryRowContext(ctx, iamOrganizationOwnerUpdateSQL,
				test.operation, bytes.Repeat([]byte{test.marker}, 32), organizationID, int64(2),
				"Denied", "denied", test.owners, test.memberships, iamIdentityOrgSchemaID,
				profile, profileSHA, iamIdentityActorID, iamIdentityPermissionID,
				actor.tokenVersion, actor.authorizationVersion)
			var id string
			var revision, clock int64
			var replayed bool
			assertIAMIdentitySQLState(t, row.Scan(&id, &revision, &clock, &replayed), test.state)
		})
	}
}
