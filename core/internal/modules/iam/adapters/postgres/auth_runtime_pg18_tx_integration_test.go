//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func beginIAMAuthRuntimeTx(
	ctx context.Context,
	database *sql.DB,
	schema string,
) (*sql.Tx, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	want := "pg_catalog, " + schema
	var bound string
	err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", want,
	).Scan(&bound)
	if err != nil || bound != want {
		_ = transaction.Rollback()
		return nil, fmt.Errorf("bind production IAM search path %s: %w", bound, err)
	}
	var first, second string
	err = transaction.QueryRowContext(ctx, `
SELECT (pg_catalog.current_schemas(false))[1]::text,
       (pg_catalog.current_schemas(false))[2]::text`).Scan(&first, &second)
	if err != nil || first != "pg_catalog" || second != schema {
		_ = transaction.Rollback()
		return nil, fmt.Errorf("IAM runtime schemas=%s,%s: %w", first, second, err)
	}
	return transaction, nil
}

type iamLoginEvidence struct {
	userID, identifierID string
	revision             int64
	sourceSHA256         []byte
	groupGeneration      int64
	iamGeneration        int64
}

func resolveIAMLoginEvidence(
	ctx context.Context,
	transaction *sql.Tx,
	fixture iamAuthRuntimeFixture,
) (iamLoginEvidence, error) {
	var evidence iamLoginEvidence
	var tenantID, kind string
	var layoutGeneration int64
	err := transaction.QueryRowContext(ctx, `
SELECT tenant_id,source_user_id,source_identifier_id,source_identifier_kind,
       source_revision,source_sha256,group_generation,iam_generation,
       layout_generation
FROM platform.iam_resolve_login_target_v2($1,'email_v1','login-directory-v1',$2)`,
		fixture.tenant.id, fixture.lookupHMAC,
	).Scan(
		&tenantID, &evidence.userID, &evidence.identifierID, &kind,
		&evidence.revision, &evidence.sourceSHA256, &evidence.groupGeneration,
		&evidence.iamGeneration, &layoutGeneration,
	)
	if err != nil || tenantID != fixture.tenant.id || kind != "email_v1" ||
		layoutGeneration != evidence.groupGeneration {
		return iamLoginEvidence{}, fmt.Errorf("resolve IAM login evidence: %w", err)
	}
	return evidence, nil
}
