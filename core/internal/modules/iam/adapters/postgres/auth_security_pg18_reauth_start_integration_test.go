//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func startIAMPasswordReauthentication(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	proof iamPasswordReauthentication,
) {
	t.Helper()
	var revision int64
	var issuedAt, expiresAt sql.NullTime
	var credentialID []byte
	var rpID sql.NullString
	var userVerification sql.NullBool
	err := transaction.QueryRowContext(ctx, `
SELECT revision,issued_at,expires_at,credential_id,rp_id,
       user_verification_required
FROM iam_reauthentication_start_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,
 'operator_console',NULL,'password_v1',NULL,'challenge-v1',$8,$9,$10,1,$11,
 false)`, proof.fixture.tenant.id, proof.binding,
		iamReauthenticationChallengeID, proof.state.userID, proof.state.sessionID,
		proof.state.tokenVersion, proof.state.authorization,
		proof.challengeHMAC, proof.contextSHA, proof.purpose, proof.origin,
	).Scan(
		&revision, &issuedAt, &expiresAt, &credentialID, &rpID,
		&userVerification,
	)
	if err != nil || revision != 1 || !issuedAt.Valid || !expiresAt.Valid ||
		len(credentialID) != 0 || rpID.Valid || userVerification.Valid ||
		!expiresAt.Time.Equal(issuedAt.Time.Add(5*60*1e9)) {
		t.Fatalf("%s reauthentication start revision=%d err=%v",
			proof.fixture.tenant.schema, revision, err)
	}
}

func readIAMPasswordReauthentication(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	proof *iamPasswordReauthentication,
) {
	t.Helper()
	var challengeRevision int64
	var attempts int16
	err := transaction.QueryRowContext(ctx, `
SELECT challenge_revision,attempt_count,password_credential_version,
       password_credential_revision
FROM iam_reauthentication_credential_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,
 'operator_console',NULL,'password_v1',$8,1,$9,$10)`,
		proof.fixture.tenant.id, proof.binding, iamReauthenticationChallengeID,
		proof.state.userID, proof.state.sessionID, proof.state.tokenVersion,
		proof.state.authorization, proof.purpose, proof.origin, proof.contextSHA,
	).Scan(
		&challengeRevision, &attempts, &proof.passwordVersion,
		&proof.passwordRevision,
	)
	if err != nil || challengeRevision != 1 || attempts != 0 ||
		proof.passwordVersion <= 0 || proof.passwordRevision <= 0 {
		t.Fatalf("%s reauthentication read revision=%d attempts=%d err=%v",
			proof.fixture.tenant.schema, challengeRevision, attempts, err)
	}
}
