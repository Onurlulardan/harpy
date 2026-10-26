package postgres

import (
	"bytes"
	"strings"
	"testing"
)

func TestIAMPasswordEnrollmentMigrationIsAtomicAndSecretFree(t *testing.T) {
	const name = "20260716014700_create_iam_identity_password_enroll_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("password enrollment migration differs by origin")
	}
	source := string(fresh)
	for _, required := range []string{
		"SECURITY DEFINER", "iam_identity_actor_assert_v1(",
		"iam_identity_replay_v1(", "iam_identity_record_v1(",
		"INSERT INTO iam_password_credentials", "argon2id_v1",
		"GRANT EXECUTE ON FUNCTION iam_identity_password_enroll_v1(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("password enrollment migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"p_password", "p_plaintext", "RAISE NOTICE"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("password enrollment migration exposes %q", forbidden)
		}
	}
}
