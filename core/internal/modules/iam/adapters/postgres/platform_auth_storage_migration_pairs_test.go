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

var platformAuthMigrationNames = []string{
	"20260716012000_create_iam_refresh_locators",
	"20260716012010_create_iam_auth_throttle_buckets",
}

var platformAuthDefaultToken = regexp.MustCompile(
	`(?m)(^|[[:space:]])DEFAULT([[:space:]]|$)`,
)

func TestIAMPlatformAuthMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range platformAuthMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMPlatformAuthMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			assertIAMPlatformAuthMigrationSource(t, fresh)
		})
	}
}

func readIAMPlatformAuthMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	path := filepath.Join(
		"..", "..", "migrations", "platform", origin,
		name+"."+origin+".sql",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func assertIAMPlatformAuthMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM platform auth migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM platform auth migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"\nGRANT ", "\nREVOKE ", "ON DELETE CASCADE", "ON DELETE SET NULL",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM platform auth migration contains forbidden token %q", forbidden)
		}
	}
	if platformAuthDefaultToken.MatchString(upper) {
		t.Fatal("IAM platform auth migration contains a DEFAULT clause")
	}
}
