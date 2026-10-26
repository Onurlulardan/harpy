package postgres

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const platformMigrationName = "20260716010000_create_iam_login_directory"

func TestIAMPlatformMigrationIsPairedAndFinal(t *testing.T) {
	fresh := readIAMPlatformMigration(t, "fresh")
	adopted := readIAMPlatformMigration(t, "v1_adopted")
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("fresh and v1_adopted platform migration content differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("platform migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(fresh))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "DEFAULT", "CREATE TEMP",
		"UNLOGGED", "PG_TEMP", "DO $$", "CREATE OR REPLACE", "GRANT ",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("platform migration contains forbidden token %q", forbidden)
		}
	}
	if bytes.Count(fresh, []byte("\n")) > 200 {
		t.Fatal("platform migration exceeds 200 lines")
	}
}

func TestIAMLoginDirectoryContractMarkers(t *testing.T) {
	source := string(readIAMPlatformMigration(t, "fresh"))
	markers := []string{
		"CREATE TABLE platform.iam_login_directory (",
		"lookup_key_id text COLLATE \"C\" NOT NULL",
		"lookup_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$'",
		"CONSTRAINT iam_login_directory_pkey PRIMARY KEY",
		"CONSTRAINT iam_login_directory_tenant_fk FOREIGN KEY (tenant_id)",
		"REFERENCES platform.tenants (id)",
		"ON UPDATE NO ACTION ON DELETE RESTRICT",
		"identifier_kind IN ('email_v1', 'portal_v1')",
		"pg_catalog.octet_length(identifier_hmac) = 32",
		"pg_catalog.octet_length(source_sha256) = 32",
		"state IN ('active', 'tombstoned')",
		"state = 'active' AND tombstoned_at IS NULL",
		"state = 'tombstoned' AND tombstoned_at IS NOT NULL",
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("login directory migration lacks marker %q", marker)
		}
	}
	if strings.Contains(source, "REFERENCES iam_users") ||
		strings.Contains(source, "normalized_value") {
		t.Fatal("platform directory must not own tenant identity or plaintext")
	}
}

func readIAMPlatformMigration(t *testing.T, origin string) []byte {
	t.Helper()
	path := filepath.Join(
		"..", "..", "migrations", "platform", origin,
		platformMigrationName+"."+origin+".sql",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
