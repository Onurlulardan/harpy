//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestIAMFormsRecordAuthorityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "forms_authority_"+origin)
			prepareIAMFormsRecordAuthority(t, ctx, database, origin)
			applyIAMFormsRecordAuthority(t, ctx, database, origin)
			applyIAMFormsLifecycleAuthority(t, ctx, database, origin)
			assertIAMFormsRecordAuthority(t, ctx, database)
			assertIAMFormsLifecycleAuthority(t, ctx, database)
			assertIAMFormsRecordImplicationAuthorization(t, ctx, database)
			assertIAMFormsRecordAuthorityRollback(t, ctx, database)
		})
	}
}

func assertIAMFormsRecordAuthorityRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	schema := iamFoundationTenants[1].schema
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var created int64
	var constraintDefinition string
	var helper sql.NullString
	err := transaction.QueryRowContext(ctx, `
SELECT (SELECT pg_catalog.count(*) FROM iam_actions
        WHERE action_key IN ('create', 'update', 'delete', 'import', 'export')),
       (SELECT pg_catalog.pg_get_constraintdef(constraint_row.oid)
        FROM pg_catalog.pg_constraint AS constraint_row
        WHERE constraint_row.conrelid = 'iam_permission_implications'::regclass
          AND constraint_row.conname = 'iam_permission_implications_kind_check'),
		pg_catalog.to_regprocedure('iam_seed_forms_record_authority_v1(name)')::text`,
	).Scan(&created, &constraintDefinition, &helper)
	if err != nil || created != 0 || helper.Valid ||
		strings.Contains(constraintDefinition, "forms_record_operation_v1") {
		t.Fatalf("drift rollback = %d/%q/%v: %v",
			created, constraintDefinition, helper, err)
	}
	var permissions, implications int64
	err = transaction.QueryRowContext(ctx, `
SELECT (SELECT pg_catalog.count(*) FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        WHERE resource.resource_key = 'forms.records'),
       (SELECT pg_catalog.count(*) FROM iam_permission_implications
        WHERE implication_kind = 'forms_record_operation_v1')`,
	).Scan(&permissions, &implications)
	if err != nil || permissions != 0 || implications != 0 {
		t.Fatalf("drift left partial authority = %d/%d: %v",
			permissions, implications, err)
	}
}
