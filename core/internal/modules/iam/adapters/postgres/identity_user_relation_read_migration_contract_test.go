package postgres

import (
	"bytes"
	"testing"
)

func TestIAMIdentityUserReadsOnlyProjectActiveRelations(t *testing.T) {
	for _, migration := range []string{
		"20260716020053_create_iam_identity_user_read_v2",
		"20260716020054_create_iam_identity_user_list_v2",
	} {
		source := readIAMIdentityMigration(t, "fresh", migration)
		for _, required := range [][]byte{
			[]byte("assignment.state = 'active'"),
			[]byte("membership.state = 'active'"),
		} {
			if !bytes.Contains(source, required) {
				t.Fatalf("%s misses active relation filter %q", migration, required)
			}
		}
		for _, forbidden := range [][]byte{
			[]byte("assignment.state <> 'revoked'"),
			[]byte("membership.state <> 'tombstoned'"),
		} {
			if bytes.Contains(source, forbidden) {
				t.Fatalf("%s exposes inactive relation through %q", migration, forbidden)
			}
		}
	}
}

func TestIAMIdentityUserReadGrantSwitchesToV2(t *testing.T) {
	source := readIAMIdentityMigration(
		t, "fresh", "20260716020055_grant_iam_identity_user_reads_v2",
	)
	for _, required := range [][]byte{
		[]byte("GRANT EXECUTE ON FUNCTION iam_identity_user_read_v2("),
		[]byte("GRANT EXECUTE ON FUNCTION iam_identity_user_list_v2("),
		[]byte("REVOKE EXECUTE ON FUNCTION iam_identity_user_read_v1("),
		[]byte("REVOKE EXECUTE ON FUNCTION iam_identity_user_list_v1("),
	} {
		if !bytes.Contains(source, required) {
			t.Fatalf("IAM user read grant switch misses %q", required)
		}
	}
}
