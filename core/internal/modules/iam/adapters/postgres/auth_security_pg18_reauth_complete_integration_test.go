//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

func completeIAMPasswordReauthentication(
	ctx context.Context, transaction *sql.Tx,
	proof iamPasswordReauthentication,
) (iamRotationProjection, error) {
	var projection iamRotationProjection
	var challengeRevision, familyRevision, sessionRevision int64
	var consumedRevision, successorRevision int64
	err := transaction.QueryRowContext(ctx, `
SELECT issued_at,challenge_revision,family_revision,session_revision,
       consumed_token_revision,successor_token_revision,locator_selector,
       locator_expected_revision,locator_revision,locator_source_sha256,
       locator_state,locator_projected_at,locator_retain_until,
       successor_selector,successor_revision,successor_source_sha256,
       successor_state,successor_issued_at,successor_projected_at,
       successor_family_expires_at,successor_retain_until
FROM iam_reauthentication_complete_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,1::bigint,0::smallint,
 $4,$5,$6,$7,$8,$9,$10,'operator_console',$11,NULL::uuid,NULL::bigint,$12,
 $13,$14,$15,NULL::uuid,0::bigint,0::bigint,$16,$17,$18,$19,'password_v1',
 0::bigint,false,false,0::bigint,$20,$21,'refresh-v1',$22)`,
		proof.fixture.tenant.id, proof.binding, iamReauthenticationChallengeID,
		proof.state.userID, proof.state.sessionID, proof.state.userRevision,
		proof.state.sessionRevision, proof.state.tokenVersion,
		proof.state.authorization, proof.state.accessModeID,
		proof.state.accessRevision, proof.state.familyID, proof.evidence.tokenID,
		proof.state.familyRevision, proof.state.tokenRevision,
		proof.passwordVersion, proof.passwordRevision, proof.purpose,
		proof.contextSHA, iamReauthenticationSuccessorID,
		proof.successorSelector, proof.successorHMAC,
	).Scan(
		&projection.issuedAt, &challengeRevision, &familyRevision,
		&sessionRevision, &consumedRevision, &successorRevision,
		&projection.locatorSelector, &projection.locatorExpected,
		&projection.locatorRevision, &projection.locatorSHA,
		&projection.locatorState, &projection.locatorProjected,
		&projection.locatorRetain, &projection.successorSelector,
		&projection.successorRevision, &projection.successorSHA,
		&projection.successorState, &projection.successorIssued,
		&projection.successorProjected, &projection.successorFamilyExpires,
		&projection.successorRetain,
	)
	projection.outcome = "rotated"
	if err == nil && (challengeRevision != 2 ||
		familyRevision != proof.state.familyRevision+1 ||
		sessionRevision != proof.state.sessionRevision+1 ||
		consumedRevision != proof.state.tokenRevision+1 || successorRevision != 1) {
		return iamRotationProjection{}, sql.ErrNoRows
	}
	return projection, err
}
