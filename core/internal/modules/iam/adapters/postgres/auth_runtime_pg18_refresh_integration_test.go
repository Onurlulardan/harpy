//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type iamRefreshEvidence struct {
	tenantID, tokenID, familyID string
	selector, sourceSHA256      []byte
	revision, groupGen, iamGen  int64
}

type iamRefreshState struct {
	tokenRevision, familyRevision, sessionRevision int64
	familyID, sessionID, userID, accessModeID      string
	userRevision, tokenVersion, authorization      int64
	accessRevision                                 int64
}

type iamRotationProjection struct {
	outcome                                 string
	issuedAt                                sql.NullTime
	locatorSelector, locatorSHA             []byte
	locatorExpected, locatorRevision        sql.NullInt64
	locatorState                            sql.NullString
	locatorProjected, locatorRetain         sql.NullTime
	successorSelector, successorSHA         []byte
	successorRevision                       sql.NullInt64
	successorState                          sql.NullString
	successorIssued, successorProjected     sql.NullTime
	successorFamilyExpires, successorRetain sql.NullTime
}

func rotateIAMRefresh(
	ctx context.Context,
	database iamPlatformGuardDatabase,
	selector []byte,
	successorID string,
	successorSelector, successorHMAC []byte,
) (iamRotationProjection, error) {
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		return iamRotationProjection{}, err
	}
	defer transaction.Rollback()
	evidence, err := resolveIAMRefreshEvidence(ctx, transaction, selector)
	if err != nil {
		return iamRotationProjection{}, err
	}
	schema, err := iamAuthSchemaForTenant(evidence.tenantID)
	if err != nil {
		return iamRotationProjection{}, err
	}
	want := "pg_catalog, " + schema
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", want); err != nil {
		return iamRotationProjection{}, err
	}
	state, err := readIAMRefreshState(ctx, transaction, evidence)
	if err != nil {
		return iamRotationProjection{}, err
	}
	projection, err := executeIAMRefreshRotation(
		ctx, transaction, state, successorID, successorSelector, successorHMAC,
	)
	if err != nil {
		return iamRotationProjection{}, err
	}
	if projection.outcome != "rejected" {
		err = advanceIAMRefreshProjection(ctx, transaction, evidence, projection)
	}
	if err == nil && projection.outcome == "rotated" {
		err = createIAMSuccessorProjection(
			ctx, transaction, evidence, state.familyID, successorID, projection,
		)
	}
	if err != nil {
		return iamRotationProjection{}, err
	}
	if err := transaction.Commit(); err != nil {
		return iamRotationProjection{}, err
	}
	return projection, nil
}

func iamAuthSchemaForTenant(tenantID string) (string, error) {
	for _, tenant := range iamFoundationTenants {
		if tenant.id == tenantID {
			return tenant.schema, nil
		}
	}
	return "", fmt.Errorf("unknown IAM auth tenant %s", tenantID)
}

func resolveIAMRefreshEvidence(
	ctx context.Context, transaction *sql.Tx, selector []byte,
) (iamRefreshEvidence, error) {
	var evidence iamRefreshEvidence
	var layoutGeneration int64
	err := transaction.QueryRowContext(ctx, `
SELECT tenant_id,source_token_id,source_family_id,source_selector,
       source_revision,source_sha256,group_generation,iam_generation,
       layout_generation
FROM platform.iam_resolve_refresh_target_v2($1)`, selector).Scan(
		&evidence.tenantID, &evidence.tokenID, &evidence.familyID,
		&evidence.selector, &evidence.revision, &evidence.sourceSHA256,
		&evidence.groupGen, &evidence.iamGen, &layoutGeneration,
	)
	if err != nil || layoutGeneration != evidence.groupGen {
		return iamRefreshEvidence{}, fmt.Errorf("resolve IAM refresh: %w", err)
	}
	return evidence, nil
}

func readIAMRefreshState(
	ctx context.Context, transaction *sql.Tx, evidence iamRefreshEvidence,
) (iamRefreshState, error) {
	var state iamRefreshState
	err := transaction.QueryRowContext(ctx, `
SELECT token_revision,family_id,family_revision,session_id,session_revision,
       user_id,user_revision,token_version,authorization_version,
       access_mode_id,access_mode_revision
FROM iam_refresh_credential_v2(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
		evidence.tokenID, evidence.familyID, evidence.selector,
		evidence.revision, evidence.sourceSHA256,
	).Scan(&state.tokenRevision, &state.familyID, &state.familyRevision,
		&state.sessionID, &state.sessionRevision, &state.userID,
		&state.userRevision, &state.tokenVersion, &state.authorization,
		&state.accessModeID, &state.accessRevision)
	return state, err
}

func projectionTime(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}
