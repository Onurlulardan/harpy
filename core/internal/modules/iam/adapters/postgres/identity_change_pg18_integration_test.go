//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
)

const iamIdentityChangeAuditWrite = `
SELECT aggregate_revision,authorization_clock,event_sequence
FROM iam_identity_change_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,'organization',$3,$4,
    'iam.identity.organization.create',
    'iam.identity.organization.create.succeeded')`

func assertIAMIdentityChangeAudit(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	aggregateID := "a8000000-0000-4000-8000-000000000001"
	for index, tenant := range iamFoundationTenants {
		actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamIdentityActorID)
		operationID := fmt.Sprintf("a%d000000-0000-4000-8000-000000000001", index+6)
		request := bytes.Repeat([]byte{byte(0x50 + index)}, 32)
		withIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
			resultClock := createIAMIdentityAuditOrganization(
				t, ctx, tx, operationID, aggregateID, request, actor,
			)
			var revision, clock, sequence int64
			err := tx.QueryRowContext(ctx, iamIdentityChangeAuditWrite,
				operationID, iamIdentityActorID, aggregateID, request,
			).Scan(&revision, &clock, &sequence)
			if err != nil || revision != 1 || clock != resultClock || sequence != 1 {
				t.Fatalf("identity change audit = %d/%d/%d error=%v",
					revision, clock, sequence, err)
			}
		})
		assertIAMIdentityChangeReplay(
			t, ctx, database, tenant.schema, operationID, aggregateID, request, actor,
		)
	}
	for _, tenant := range iamFoundationTenants {
		var count int
		mustQueryIAMTenantValue(t, ctx, database, tenant.schema,
			`SELECT pg_catalog.count(*) FROM iam_identity_change_audit
             WHERE aggregate_id='a8000000-0000-4000-8000-000000000001'`, &count)
		if count != 1 {
			t.Fatalf("tenant %s identity audit count=%d", tenant.schema, count)
		}
	}
}

func createIAMIdentityAuditOrganization(
	t *testing.T, ctx context.Context, tx *sql.Tx,
	operationID, aggregateID string, request []byte, actor iamIdentityActor,
) int64 {
	t.Helper()
	profile, digest := []byte(`{}`), sha256.Sum256([]byte(`{}`))
	var id string
	var revision, clock int64
	var replayed bool
	name, slug := "Audited "+aggregateID[:8], "audited-"+aggregateID[:8]
	err := tx.QueryRowContext(ctx, iamIdentityOrganizationCreate,
		operationID, request, aggregateID, name, slug,
		iamIdentityOrgSchemaID, profile, digest[:], iamIdentityActorID,
		iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
	).Scan(&id, &revision, &clock, &replayed)
	if err != nil || id != aggregateID || revision != 1 || replayed {
		t.Fatalf("identity audit organization create=%s/%d/%d/%t error=%v",
			id, revision, clock, replayed, err)
	}
	return clock
}

func assertIAMIdentityChangeReplay(
	t *testing.T, ctx context.Context, database *sql.DB, schema,
	operationID, aggregateID string, request []byte, actor iamIdentityActor,
) {
	t.Helper()
	withIAMIdentityRuntimeTx(t, ctx, database, schema, func(tx *sql.Tx) {
		profile, digest := []byte(`{}`), sha256.Sum256([]byte(`{}`))
		var id string
		var revision, clock int64
		var replayed bool
		name, slug := "Audited "+aggregateID[:8], "audited-"+aggregateID[:8]
		err := tx.QueryRowContext(ctx, iamIdentityOrganizationCreate,
			operationID, request, aggregateID, name, slug,
			iamIdentityOrgSchemaID, profile, digest[:], iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
		).Scan(&id, &revision, &clock, &replayed)
		if err != nil || !replayed {
			t.Fatalf("identity change replay=%t error=%v", replayed, err)
		}
	})
}
