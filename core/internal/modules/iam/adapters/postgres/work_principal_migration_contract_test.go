package postgres

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMWorkPrincipalMigrationsUseCurrentAccessModeAuthority(t *testing.T) {
	for _, name := range []string{
		"20260801010000_create_iam_work_principal_evidence_v1",
		"20260801010001_create_iam_work_available_users_v1",
	} {
		fresh := readIAMIdentityMigration(t, "fresh", name)
		if len(fresh) == 0 {
			t.Fatal("fresh migration is empty")
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect %s: %v", name, err)
		}
		source := string(fresh)
		for _, marker := range []string{
			"iam_user_access_modes",
			"access_mode.mode = 'operator_console'",
			"access_mode.state = 'active'",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks current access authority %q", name, marker)
			}
		}
		if strings.Contains(source, ".account_type") {
			t.Fatalf("%s reads removed iam_users.account_type", name)
		}
	}
}
