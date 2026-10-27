//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func createIAMExpiredFixtureLocator(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) {
	t.Helper()
	var issuedAt, projectedAt, familyExpiresAt, retainUntil time.Time
	queryIAMAuthState(t, ctx, database.migrator, fixture.tenant.schema, `
SELECT issued_at,updated_at,family_absolute_expires_at,retain_until
FROM iam_refresh_tokens
WHERE id='64000000-0000-4000-8000-000000000002'`,
		&issuedAt, &projectedAt, &familyExpiresAt, &retainUntil)
	_, err := database.runtime.ExecContext(ctx, iamCreateLocatorSQL,
		fixture.expiredSelect, fixture.tenant.id, iamExpiredTokenID,
		iamExpiredFamilyID, int64(1), int64(1), int64(1),
		fixture.expiredSHA, "active", issuedAt, projectedAt,
		familyExpiresAt, retainUntil)
	if err != nil {
		t.Fatalf("create %s expired IAM locator: %v", fixture.tenant.schema, err)
	}
}

func assertIAMAuthIdleExpiry(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) {
	t.Helper()
	projection, err := rotateIAMRefresh(
		ctx, database, fixture.expiredSelect,
		"65000000-0000-4000-8000-000000000002",
		iamBytes(fixture.expiredSelect[0]+1, 16),
		iamBytes(fixture.expiredSelect[0]+2, 32),
	)
	if err != nil || projection.outcome != "expired" ||
		projection.locatorState.String != "expired" {
		t.Fatalf("%s IAM idle expiry=%s/%s: %v", fixture.tenant.schema,
			projection.outcome, projection.locatorState.String, err)
	}
	var tokenState, familyState, sessionState, locatorState string
	queryIAMAuthState(t, ctx, database.migrator, fixture.tenant.schema, `
SELECT token.state,family.state,session.state
FROM iam_refresh_tokens AS token
JOIN iam_refresh_families AS family ON family.id=token.family_id
JOIN iam_sessions AS session ON session.id=family.session_id
WHERE token.id='64000000-0000-4000-8000-000000000002'`,
		&tokenState, &familyState, &sessionState)
	if err := database.migrator.QueryRowContext(ctx, `
SELECT source_token_state FROM platform.iam_refresh_locators WHERE selector=$1`,
		fixture.expiredSelect).Scan(&locatorState); err != nil {
		t.Fatal(err)
	}
	if tokenState != "expired" || familyState != "active" ||
		sessionState != "active" || locatorState != "expired" {
		t.Fatalf("%s idle state=%s/%s/%s/%s", fixture.tenant.schema,
			tokenState, familyState, sessionState, locatorState)
	}
}
