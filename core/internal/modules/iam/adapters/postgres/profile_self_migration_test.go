package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamSelfProfileMigrationNames = []string{
	"20260716015910_create_iam_self_profile_read_v1",
	"20260716015911_create_iam_self_profile_update_v1",
	"20260716015912_grant_iam_self_profile_runtime_v1",
}

func TestIAMSelfProfileMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamSelfProfileMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("self profile migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("inspect self profile migration: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE",
				"CREATE TEMP", "UNLOGGED", "PG_TEMP", "PUBLIC.",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("self profile migration contains %q", forbidden)
				}
			}
		})
	}
}

func TestIAMSelfProfileRoutinesKeepActorCASAndNarrowGrant(t *testing.T) {
	read := string(readIAMIdentityMigration(t, "fresh", iamSelfProfileMigrationNames[0]))
	update := string(readIAMIdentityMigration(t, "fresh", iamSelfProfileMigrationNames[1]))
	grant := string(readIAMIdentityMigration(t, "fresh", iamSelfProfileMigrationNames[2]))
	for _, marker := range []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"actor.token_version = p_expected_token_version",
		"actor.authorization_version = p_expected_authorization_version",
	} {
		if !strings.Contains(read, marker) || !strings.Contains(update, marker) {
			t.Fatalf("self profile authority lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"p_expected_source_schema_id", "p_expected_revision",
		"iam_authorization_operation_receipts", "iam_identity_profile_assert_v1",
	} {
		if !strings.Contains(update, marker) {
			t.Fatalf("self profile update lacks %q", marker)
		}
	}
	if strings.Count(grant, "GRANT EXECUTE ON FUNCTION") != 2 ||
		strings.Contains(grant, "GRANT SELECT") || strings.Contains(grant, "GRANT UPDATE") {
		t.Fatal("self profile grant is not execute-only")
	}
}
