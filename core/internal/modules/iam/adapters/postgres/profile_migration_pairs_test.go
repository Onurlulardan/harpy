package postgres

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var profileMigrationNames = []string{
	"20260716011000_create_iam_profile_schema_versions",
	"20260716011100_create_iam_profile_fields",
	"20260716011200_create_iam_profile_field_revisions",
}

func TestIAMProfileMigrationsStayPairedAndFinal(t *testing.T) {
	for _, name := range profileMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readProfileMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			assertFinalProfileMigrationSource(t, fresh)
		})
	}
}

func readProfileMigration(t *testing.T, origin, name string) []byte {
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

func assertFinalProfileMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n")) > 200 {
		t.Fatal("profile migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("profile migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "DEFAULT",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP", "PUBLIC.",
		"DO $$", "CREATE OR REPLACE",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("profile migration contains forbidden token %q", forbidden)
		}
	}
}
