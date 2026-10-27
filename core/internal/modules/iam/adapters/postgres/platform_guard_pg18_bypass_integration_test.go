//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertIAMPlatformGuardBypasses(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	assertIAMDirectGrantStillFenced(t, ctx, database)
	assertIAMAlwaysTriggerInReplicaMode(t, ctx, database)
	assertIAMThrottleGuardRejectsInvalidReset(t, ctx, database)
	assertIAMPlatformSearchPathAndAdopter(t, ctx, database)
}

const iamDirectLocatorInsertSQL = `INSERT INTO platform.iam_refresh_locators(
 selector,tenant_id,token_id,family_id,scope_kind,scope_id,
 authority_group_id,authority_group_generation,authority_domain,cutover_generation,
 source_revision,source_sha256,source_token_state,issued_at,projected_at,
 family_absolute_expires_at,retain_until)
VALUES($1,$2,$3,$4,'tenant',$2,'iam_system_resources',$5,'iam',$5,
1,$6,'active',$7,$8,$9,$10)`

func directIAMLocatorInsert(
	ctx context.Context, database iamPlatformExecer, selector []byte,
) error {
	now := time.Now().UTC()
	_, err := database.ExecContext(ctx, iamDirectLocatorInsertSQL,
		selector, iamGuardTenantA, uuid.New(), uuid.New(), int64(2), iamBytes(61, 32),
		now.Add(-time.Minute), now, now.Add(48*time.Hour), now.Add(73*time.Hour))
	return err
}

func assertIAMDirectGrantStillFenced(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	_, err := database.migrator.ExecContext(ctx, `GRANT INSERT,UPDATE ON
platform.iam_refresh_locators TO harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
	assertIAMSQLState(t, directIAMLocatorInsert(ctx, database.runtime,
		iamBytes(62, 16)), "42501")
	_, err = database.migrator.ExecContext(ctx, `REVOKE INSERT,UPDATE ON
platform.iam_refresh_locators FROM harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.migrator.ExecContext(ctx, `UPDATE platform.iam_refresh_locators
SET retain_until=retain_until+interval '1 hour' WHERE selector=$1`, iamBytes(17, 16))
	assertIAMSQLState(t, err, "42501")
}

func assertIAMAlwaysTriggerInReplicaMode(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	connection, err := database.admin.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "SET session_replication_role=replica"); err != nil {
		t.Fatal(err)
	}
	assertIAMSQLState(t, directIAMLocatorInsert(ctx, connection, iamBytes(63, 16)), "42501")
	if _, err := connection.ExecContext(ctx, "SET session_replication_role=origin"); err != nil {
		t.Fatal(err)
	}
}

const iamThrottleProbeSQL = `
CREATE FUNCTION platform.test_iam_throttle_guard_probe_v1(p_key text,p_mode text)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $f$
BEGIN
 IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
    CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN RAISE EXCEPTION USING ERRCODE='42501'; END IF;
 IF p_mode OPERATOR(pg_catalog.=) 'clear' THEN
  UPDATE platform.iam_auth_throttle_buckets AS bucket
  SET policy_version=bucket.policy_version+1,
      window_start=bucket.last_attempt_at+interval '1 microsecond',
      window_end=bucket.window_end+interval '1 microsecond',attempt_count=1,
      blocked_until=NULL,last_attempt_at=bucket.last_attempt_at+interval '2 microseconds',
      revision=bucket.revision+1,retain_until=bucket.retain_until+interval '1 microsecond'
  WHERE bucket.key_id OPERATOR(pg_catalog.=) p_key;
 ELSE
  UPDATE platform.iam_auth_throttle_buckets AS bucket
  SET window_start=bucket.last_attempt_at+interval '1 microsecond',attempt_count=1,
      last_attempt_at=bucket.last_attempt_at+interval '2 microseconds',
      revision=bucket.revision+1
  WHERE bucket.key_id OPERATOR(pg_catalog.=) p_key;
 END IF;
END;$f$;
REVOKE ALL ON FUNCTION platform.test_iam_throttle_guard_probe_v1(text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.test_iam_throttle_guard_probe_v1(text,text)
TO harpy_runtime;`

func assertIAMThrottleGuardRejectsInvalidReset(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	for range iamGuardIdentifierThreshold {
		_, _, _, err := recordIAMTenantAttempt(ctx, database.runtime,
			iamGuardTenantA, "guard-probe", iamBytes(64, 32), 2,
			iamGuardPolicyVersion, iamGuardAuthWindow,
			iamGuardIdentifierThreshold, iamGuardAuthBlock)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.migrator.ExecContext(ctx, iamThrottleProbeSQL); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"clear", "early"} {
		_, err := database.runtime.ExecContext(ctx,
			"SELECT platform.test_iam_throttle_guard_probe_v1($1,$2)",
			"guard-probe", mode)
		assertIAMSQLState(t, err, "55000")
	}
	if _, err := database.migrator.ExecContext(ctx,
		"DROP FUNCTION platform.test_iam_throttle_guard_probe_v1(text,text)"); err != nil {
		t.Fatal(err)
	}
}

func assertIAMPlatformSearchPathAndAdopter(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	connection, err := database.runtime.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "SET search_path=pg_temp,public"); err != nil {
		t.Fatal(err)
	}
	var tenant string
	var group, iam, layout int64
	err = connection.QueryRowContext(ctx,
		"SELECT * FROM platform.iam_resolve_tenant_target_v1($1)", iamGuardTenantA).
		Scan(&tenant, &group, &iam, &layout)
	if err != nil || tenant != iamGuardTenantA || group != 2 || iam != 2 || layout != 2 {
		t.Fatalf("poisoned search_path target=%s/%d/%d/%d err=%v",
			tenant, group, iam, layout, err)
	}
	err = database.adopter.QueryRowContext(ctx,
		"SELECT * FROM platform.iam_resolve_tenant_target_v1($1)", iamGuardTenantA).
		Scan(&tenant, &group, &iam, &layout)
	assertIAMSQLState(t, err, "42501")
	var count int
	err = database.adopter.QueryRowContext(ctx,
		"SELECT count(*) FROM platform.iam_refresh_locators").Scan(&count)
	assertIAMSQLState(t, err, "42501")
}
