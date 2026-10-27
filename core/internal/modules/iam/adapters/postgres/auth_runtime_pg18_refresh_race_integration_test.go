//go:build integration

package postgres

import (
	"context"
	"fmt"
	"sort"
	"testing"
)

func assertIAMAuthRefreshConcurrency(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) {
	t.Helper()
	outcomes := make(chan string, 2)
	errors := make(chan error, 2)
	for index := byte(1); index <= 2; index++ {
		index := index
		go func() {
			projection, err := rotateIAMRefresh(
				ctx, database, fixture.selector,
				fmt.Sprintf("65000000-0000-4000-8000-%012d", index),
				iamBytes(fixture.selector[0]+0x10+index, 16),
				iamBytes(fixture.secretHMAC[0]+0x10+index, 32),
			)
			outcomes <- projection.outcome
			errors <- err
		}()
	}
	results := []string{<-outcomes, <-outcomes}
	for range 2 {
		if err := <-errors; err != nil {
			t.Fatalf("%s IAM refresh concurrency: %v", fixture.tenant.schema, err)
		}
	}
	sort.Strings(results)
	if results[0] != "reused" || results[1] != "rotated" {
		t.Fatalf("%s IAM refresh outcomes=%v", fixture.tenant.schema, results)
	}
	assertIAMAuthReuseState(t, ctx, database, fixture)
}

func assertIAMAuthReuseState(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) {
	t.Helper()
	var family, session string
	var tokenVersion, activeTokens, successorCount int
	queryIAMAuthState(t, ctx, database.migrator, fixture.tenant.schema, `
SELECT family.state,session.state,principal.token_version,
 (SELECT count(*) FROM iam_refresh_tokens AS active
  WHERE active.family_id=family.id AND active.state='active'),
 (SELECT count(*) FROM iam_refresh_tokens AS child
  WHERE child.parent_id='64000000-0000-4000-8000-000000000001')
FROM iam_refresh_families AS family
JOIN iam_sessions AS session ON session.id=family.session_id
JOIN iam_users AS principal ON principal.id=family.user_id
WHERE family.id='63000000-0000-4000-8000-000000000001'`,
		&family, &session, &tokenVersion, &activeTokens, &successorCount)
	if family != "compromised" || session != "revoked" || tokenVersion != 2 ||
		activeTokens != 0 || successorCount != 1 {
		t.Fatalf("%s IAM reuse=%s/%s/%d/%d/%d", fixture.tenant.schema,
			family, session, tokenVersion, activeTokens, successorCount)
	}
}
