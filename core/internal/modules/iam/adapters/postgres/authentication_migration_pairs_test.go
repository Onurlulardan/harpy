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

var iamAuthenticationMigrationNames = []string{
	"20260716012300_create_iam_authenticators",
	"20260716012400_create_iam_webauthn_credentials",
	"20260716012500_create_iam_totp_authenticators",
	"20260716012600_create_iam_recovery_codes",
	"20260716012700_create_iam_auth_challenges",
}

var authenticationDefaultToken = regexp.MustCompile(
	`(?m)(^|[[:space:]])DEFAULT([[:space:]]|$)`,
)

func TestIAMAuthenticationMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamAuthenticationMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMAuthenticationMigration(t, "fresh", name)
			adopted := readIAMAuthenticationMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migration content differs")
			}
			assertIAMAuthenticationMigrationSource(t, fresh)
		})
	}
}

func readIAMAuthenticationMigration(t *testing.T, origin, name string) []byte {
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

func assertIAMAuthenticationMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM authentication migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM authentication migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"PUBLIC.", "\nGRANT ", "\nREVOKE ", "ON DELETE CASCADE",
		"ON DELETE SET NULL", "PRIVATE_KEY", "PLAINTEXT", "RAW_SECRET",
		"SECRET TEXT", "CHALLENGE TEXT", "CODE TEXT",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM authentication migration contains forbidden token %q", forbidden)
		}
	}
	if authenticationDefaultToken.MatchString(upper) {
		t.Fatal("IAM authentication migration contains a DEFAULT clause")
	}
}
