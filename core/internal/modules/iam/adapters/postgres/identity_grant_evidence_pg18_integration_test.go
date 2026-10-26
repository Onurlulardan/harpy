//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
)

const iamIdentityGrantEvidenceMigration = "20260730113400_accept_iam_identity_grant_evidence_v1"

func applyIAMIdentityGrantEvidenceAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	names := []string{
		"20260716013800_create_iam_profile_actor_assert_v1",
		"20260716015800_create_iam_v1_actor_assertion_authority_v1",
		"20260716015801_grant_iam_v1_actor_assertion_authority_v1",
		"20260727150004_accept_iam_protected_rbac_evidence_v2",
		iamIdentityGrantEvidenceMigration,
	}
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func assertIAMIdentityGrantEvidenceAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const actorID = "51000000-0000-4000-8000-000000000002"
	const organizationID = "a2000000-0000-4000-8000-000000000099"
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", actorID)
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		profile := []byte(`{}`)
		profileDigest := sha256.Sum256(profile)
		var id string
		var revision, clock int64
		var replayed bool
		err := transaction.QueryRowContext(ctx, iamIdentityOrganizationCreate,
			"a3000000-0000-4000-8000-000000000099",
			bytes.Repeat([]byte{0x39}, 32), organizationID,
			"Grant Evidence Organization", "grant-evidence-organization",
			iamIdentityOrgSchemaID, profile, profileDigest[:],
			actorID, iamRBACTenantAllID,
			actor.tokenVersion, actor.authorizationVersion,
		).Scan(&id, &revision, &clock, &replayed)
		if err != nil || id != organizationID || revision != 1 || replayed {
			t.Fatalf("identity mutation rejected protected tenant grant evidence: %v", err)
		}
		var records int
		err = transaction.QueryRowContext(ctx, `
SELECT count(*) FROM iam_identity_organization_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
			organizationID, actorID, iamRBACTenantAllID,
			actor.tokenVersion, actor.authorizationVersion,
		).Scan(&records)
		if err != nil || records != 1 {
			t.Fatalf("identity query rejected protected tenant grant evidence: %v", err)
		}
	})
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		profile := []byte(`{}`)
		profileDigest := sha256.Sum256(profile)
		var id string
		var revision, clock int64
		var replayed bool
		err := transaction.QueryRowContext(ctx, iamIdentityOrganizationCreate,
			"a3000000-0000-4000-8000-000000000098",
			bytes.Repeat([]byte{0x38}, 32),
			"a2000000-0000-4000-8000-000000000098",
			"Stale Grant Evidence", "stale-grant-evidence",
			iamIdentityOrgSchemaID, profile, profileDigest[:],
			actorID, iamRBACTenantAllID,
			actor.tokenVersion, actor.authorizationVersion+1,
		).Scan(&id, &revision, &clock, &replayed)
		assertIAMIdentitySQLState(t, err, "42501")
	})
}
