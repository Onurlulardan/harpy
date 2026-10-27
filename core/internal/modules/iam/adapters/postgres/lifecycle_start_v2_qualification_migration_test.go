package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamLifecycleStartV2QualificationMigrations = []string{
	"20260725124600_qualify_iam_password_reset_start_v2",
	"20260725124601_qualify_iam_email_verification_start_v2",
	"20260725124602_qualify_iam_portal_setup_start_v2",
}

var iamLifecycleKeyContractMigrations = []string{
	"20260725124700_repair_iam_one_time_key_contract",
	"20260725124701_repair_iam_password_reset_complete_key_contract",
}

func TestIAMLifecycleStartV2QualificationMigrationsStayPaired(t *testing.T) {
	for _, name := range iamLifecycleStartV2QualificationMigrations {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("lifecycle qualification migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("inspect lifecycle qualification migration: %v", err)
			}
			source := string(fresh)
			for _, marker := range []string{
				"CREATE OR REPLACE FUNCTION", "SELECT started.result_id",
				"started.subject_user_id", "started.result_revision",
				"started.token_version", "started.authorization_version",
				"started.replayed", ") AS started;", "REVOKE ALL ON FUNCTION",
			} {
				if !strings.Contains(source, marker) {
					t.Fatalf("lifecycle qualification is missing %q", marker)
				}
			}
		})
	}
}

func TestIAMLifecycleKeyContractMigrationsStayPaired(t *testing.T) {
	for _, name := range iamLifecycleKeyContractMigrations {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("lifecycle key repair exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("inspect lifecycle key repair: %v", err)
			}
			source := string(fresh)
			if !strings.Contains(source, "NOT iam_hmac_key_id_valid_v1(p_hash_key_id)") ||
				strings.Contains(source, "p_hash_key_id !~") {
				t.Fatal("lifecycle key repair does not use the canonical helper")
			}
		})
	}
}
