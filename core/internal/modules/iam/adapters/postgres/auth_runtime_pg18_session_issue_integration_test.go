//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func issueIAMRuntimeSession(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture iamAuthRuntimeFixture,
	used iamLoginState,
	login iamLoginState,
) (time.Time, time.Time, time.Time, time.Time) {
	t.Helper()
	var sessionRevision, familyRevision, tokenRevision int64
	var issuedAt, idle, absolute, retain time.Time
	err := transaction.QueryRowContext(ctx, `
SELECT session_revision,family_revision,token_revision,issued_at,
       idle_expires_at,absolute_expires_at,retain_until
FROM iam_session_issue_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,
 'operator_console',$6,NULL,NULL,$7,'desktop','chromium','windows',
 'network-v1',$8,$9,$10,$11,'refresh-v1',$12)`,
		iamUserOneID, used.userRevision, used.tokenVersion,
		used.authorizationVersion, login.accessModeID, login.accessRevision,
		iamAuthSessionID, iamBytes(0xa5, 32), iamAuthFamilyID, iamAuthTokenID,
		fixture.selector, fixture.secretHMAC,
	).Scan(&sessionRevision, &familyRevision, &tokenRevision,
		&issuedAt, &idle, &absolute, &retain)
	if err != nil || sessionRevision != 1 || familyRevision != 1 || tokenRevision != 1 {
		_ = transaction.Rollback()
		t.Fatalf("%s IAM session issue revisions=%d/%d/%d err=%v",
			fixture.tenant.schema, sessionRevision, familyRevision, tokenRevision, err)
	}
	return issuedAt.UTC(), idle.UTC(), absolute.UTC(), retain.UTC()
}
