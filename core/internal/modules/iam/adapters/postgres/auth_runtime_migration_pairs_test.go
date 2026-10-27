package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamAuthRuntimeMigrationNames = []string{
	"20260716013300_create_iam_login_credential_reader_v1",
	"20260716013310_create_iam_password_use_v1",
	"20260716013320_create_iam_session_issue_v1",
	"20260716013330_create_iam_refresh_credential_reader_v1",
	"20260716013349_create_iam_refresh_decide_v2",
	"20260716013350_create_iam_refresh_reuse_v2",
	"20260716013351_create_iam_refresh_expire_v2",
	"20260716013352_create_iam_refresh_rotate_active_v2",
	"20260716013353_create_iam_refresh_rotate_v2",
	"20260716013354_grant_iam_auth_runtime_v2",
	"20260716013355_create_iam_refresh_locator_source_sha256_v1",
	"20260716013356_create_iam_refresh_credential_reader_v2",
	"20260716013357_create_iam_refresh_rotate_v3",
	"20260716013358_grant_iam_auth_runtime_v3",
	"20260716013359_create_iam_reauthentication_digests_v1",
	"20260716014200_create_iam_reauthentication_start_v1",
	"20260716014201_create_iam_reauthentication_credential_v1",
	"20260716014202_create_iam_reauthentication_reject_v1",
	"20260716014203_create_iam_reauthentication_proof_apply_v1",
	"20260716014204_create_iam_reauthentication_rotate_active_v1",
	"20260716014205_create_iam_reauthentication_complete_v1",
	"20260716014206_create_iam_session_revoke_v1",
	"20260716014207_create_iam_recovery_credential_v1",
	"20260716014208_create_iam_recovery_consume_v2",
	"20260716014209_grant_iam_auth_runtime_v4",
}

func TestIAMAuthRuntimeMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamAuthRuntimeMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted IAM auth runtime differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM auth runtime migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM auth runtime migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			forbiddenTokens := []string{
				"IF NOT EXISTS", "CREATE OR REPLACE", "-- +GOOSE DOWN",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"DISABLE TRIGGER", "ON DELETE CASCADE",
			}
			if !strings.Contains(name, "grant_iam_auth_runtime_v") {
				forbiddenTokens = append(forbiddenTokens, "\nGRANT ")
			}
			for _, forbidden := range forbiddenTokens {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("IAM auth runtime contains forbidden token %q", forbidden)
				}
			}
		})
	}
}
