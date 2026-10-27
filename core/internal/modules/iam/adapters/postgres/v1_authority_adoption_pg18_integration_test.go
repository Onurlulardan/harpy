//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"
)

const v1AdoptionOperation = "71000000-0000-4000-8000-000000000009"

func TestV1ProtectedAuthorityAdoptionAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "v1_adoption_"+origin)
			prepareV1AuthorityAdoption(t, ctx, database, origin)
			var direct bool
			err := database.adopter.QueryRowContext(ctx, `SELECT has_function_privilege(
'harpy_adopter',
'platform.iam_v1_adopt_protected_authority_v1(uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text)',
'EXECUTE')`).Scan(&direct)
			if err != nil || direct {
				t.Fatalf("direct V1 authority owner routine privilege=%v err=%v", direct, err)
			}
			if _, err := database.adopter.ExecContext(ctx, `SET application_name =
'harpy-cutover/71000000-0000-4000-8000-000000000010'`); err != nil {
				t.Fatal(err)
			}
			batch, digest := v1AuthorityAdoptionBatch(t)
			count, err := callV1AuthorityAdoption(ctx, database.adopter, batch, digest)
			if err != nil || count != 1 {
				t.Fatalf("first V1 authority adoption count=%d err=%v", count, err)
			}
			count, err = callV1AuthorityAdoption(ctx, database.adopter, batch, digest)
			if err != nil || count != 1 {
				t.Fatalf("exact V1 authority replay count=%d err=%v", count, err)
			}
			assertV1AuthorityAdoptionReceipt(t, ctx, database)
		})
	}
}

func assertV1AuthorityAdoptionReceipt(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var subjects, owners, receipts int
	var authorizationVersion int64
	err := database.admin.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.iam_platform_subjects),
(SELECT count(*) FROM platform.iam_platform_owner_policies),
(SELECT count(*) FROM platform.iam_v1_protected_authority_adoption_receipts),
(SELECT authorization_version FROM platform.iam_v1_protected_authority_adoption_receipts)`).
		Scan(&subjects, &owners, &receipts, &authorizationVersion)
	if err != nil || subjects != 1 || owners != 1 || receipts != 1 ||
		authorizationVersion != 7 {
		t.Fatalf("adoption counts=%d/%d/%d version=%d err=%v",
			subjects, owners, receipts, authorizationVersion, err)
	}
	_, err = database.admin.ExecContext(ctx, `UPDATE
platform.iam_v1_protected_authority_adoption_receipts SET plan_sha256=decode(repeat('c',64),'hex')`)
	if err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("mutable V1 authority receipt err=%v", err)
	}
}
