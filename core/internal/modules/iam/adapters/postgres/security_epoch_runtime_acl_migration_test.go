package postgres

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMSecurityEpochRuntimeACLUsesCurrentRole(t *testing.T) {
	const name = "20260725125100_repair_iam_security_epoch_runtime_acl"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("security epoch ACL migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, routine := range []string{
		"iam_security_epoch_current_v1", "iam_session_issue_v2",
		"iam_refresh_credential_v3", "iam_refresh_rotate_v4",
		"iam_current_session_read_v3", "iam_security_epoch_assert_v1",
	} {
		if !strings.Contains(source, "GRANT EXECUTE ON FUNCTION "+routine) {
			t.Fatalf("security epoch ACL repair lacks current routine %q", routine)
		}
	}
	if strings.Count(source, "TO harpy_runtime;") != 6 {
		t.Fatal("security epoch ACL repair does not grant exactly six current routines")
	}
	if strings.Count(source, "FROM harpy_v2_runtime;") != 7 {
		t.Fatal("security epoch ACL repair does not revoke every session-user grant")
	}
}
