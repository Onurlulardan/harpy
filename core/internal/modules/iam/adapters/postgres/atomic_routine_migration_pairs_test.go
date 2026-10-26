package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamAtomicRoutineMigrationNames = []string{
	"20260716012950_create_iam_refresh_rotate_v1",
	"20260716012951_create_iam_one_time_consume_v1",
	"20260716012952_create_iam_invitation_accept_v1",
	"20260716012953_create_iam_challenge_resolve_v1",
	"20260716012954_create_iam_recovery_consume_v1",
	"20260716012955_create_iam_portal_setup_complete_v1",
	"20260716012956_grant_iam_atomic_routines_v1",
}

func TestIAMAtomicRoutineMigrationsStayPairedAndBounded(t *testing.T) {
	for _, name := range iamAtomicRoutineMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted atomic routine differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM atomic routine migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM atomic routine migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"-- +GOOSE DOWN", "CREATE OR REPLACE", "\nGRANT SELECT",
				"\nGRANT INSERT", "\nGRANT UPDATE", "\nGRANT DELETE",
				"GRANT ALL ON TABLE", " FROM PUBLIC;\nGRANT",
				"EXECUTE COMMAND", "PG_TEMP", "PUBLIC.",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("IAM atomic routine contains forbidden token %q", forbidden)
				}
			}
		})
	}
}
