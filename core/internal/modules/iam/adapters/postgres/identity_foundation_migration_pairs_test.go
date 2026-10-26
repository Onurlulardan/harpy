package postgres

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamIdentityMigrationNames = []string{
	"20260716011300_create_iam_users",
	"20260716011400_create_iam_user_identifiers",
	"20260716011500_create_iam_organizations",
	"20260716011600_create_iam_password_credentials",
	"20260716011700_create_iam_organization_memberships",
	"20260716011800_create_iam_roles",
	"20260716011900_create_iam_user_role_assignments",
	"20260716012100_correct_iam_human_access",
	"20260716012120_create_iam_portal_subject_bindings",
	"20260716012130_create_iam_portal_access_setups",
}

var sqlDefaultToken = regexp.MustCompile(`(?m)(^|[[:space:]])DEFAULT([[:space:]]|$)`)

func TestIAMIdentityMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamIdentityMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migration content differs")
			}
			assertIAMIdentityMigrationSource(t, fresh)
		})
	}
}

func readIAMIdentityMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	path := filepath.Join(
		"..", "..", "migrations", "tenant", origin,
		name+"."+origin+".sql",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func assertIAMIdentityMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM identity migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM identity migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$",
		"CREATE OR REPLACE", "CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED",
		"PG_TEMP", "PUBLIC.", "\nGRANT ", "\nREVOKE ",
		"ON DELETE CASCADE", "ON DELETE SET NULL",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM identity migration contains forbidden token %q", forbidden)
		}
	}
	if sqlDefaultToken.MatchString(upper) {
		t.Fatal("IAM identity migration contains a DEFAULT clause")
	}
}
