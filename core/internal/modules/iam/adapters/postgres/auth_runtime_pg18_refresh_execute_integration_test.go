//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

func executeIAMRefreshRotation(
	ctx context.Context,
	transaction *sql.Tx,
	state iamRefreshState,
	successorID string,
	successorSelector, successorHMAC []byte,
) (iamRotationProjection, error) {
	var projection iamRotationProjection
	err := transaction.QueryRowContext(ctx, `
SELECT outcome,issued_at,locator_selector,locator_expected_revision,
       locator_revision,locator_source_sha256,locator_state,
       locator_projected_at,locator_retain_until,successor_selector,
       successor_revision,successor_source_sha256,successor_state,
       successor_issued_at,successor_projected_at,
       successor_family_expires_at,successor_retain_until
FROM iam_refresh_rotate_v3(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8,$9,
 $10,$11,$12,NULL,NULL,$13,$14,'refresh-v1',$15)`,
		iamAuthTokenIDFor(state.familyID), state.familyID, state.tokenRevision,
		state.familyRevision, state.sessionID, state.sessionRevision,
		state.userID, state.userRevision, state.tokenVersion, state.authorization,
		state.accessModeID, state.accessRevision, successorID, successorSelector,
		successorHMAC,
	).Scan(
		&projection.outcome, &projection.issuedAt, &projection.locatorSelector,
		&projection.locatorExpected, &projection.locatorRevision,
		&projection.locatorSHA, &projection.locatorState,
		&projection.locatorProjected, &projection.locatorRetain,
		&projection.successorSelector, &projection.successorRevision,
		&projection.successorSHA, &projection.successorState,
		&projection.successorIssued, &projection.successorProjected,
		&projection.successorFamilyExpires, &projection.successorRetain,
	)
	return projection, err
}

func iamAuthTokenIDFor(familyID string) string {
	if familyID == iamExpiredFamilyID {
		return iamExpiredTokenID
	}
	return iamAuthTokenID
}

func advanceIAMRefreshProjection(
	ctx context.Context,
	transaction *sql.Tx,
	evidence iamRefreshEvidence,
	projection iamRotationProjection,
) error {
	_, err := transaction.ExecContext(ctx, iamAdvanceLocatorSQL,
		projection.locatorSelector, evidence.tenantID,
		evidence.groupGen, evidence.iamGen,
		projection.locatorExpected.Int64, projection.locatorRevision.Int64,
		projection.locatorSHA, projection.locatorState.String,
		projectionTime(projection.locatorProjected),
		projectionTime(projection.locatorRetain))
	return err
}

func createIAMSuccessorProjection(
	ctx context.Context,
	transaction *sql.Tx,
	evidence iamRefreshEvidence,
	familyID, successorID string,
	projection iamRotationProjection,
) error {
	_, err := transaction.ExecContext(ctx, iamCreateLocatorSQL,
		projection.successorSelector, evidence.tenantID, successorID, familyID,
		evidence.groupGen, evidence.iamGen, projection.successorRevision.Int64,
		projection.successorSHA, projection.successorState.String,
		projectionTime(projection.successorIssued),
		projectionTime(projection.successorProjected),
		projectionTime(projection.successorFamilyExpires),
		projectionTime(projection.successorRetain))
	return err
}
