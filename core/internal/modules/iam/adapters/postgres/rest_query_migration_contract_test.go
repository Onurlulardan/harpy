package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamRestQueryMigrationNames = []string{
	"20260716014100_create_iam_rest_query_actor_versions_v1",
	"20260716014101_create_iam_rest_query_users_v1",
	"20260716014102_create_iam_rest_query_organizations_v1",
	"20260716014103_create_iam_rest_query_roles_v1",
	"20260716014104_create_iam_rest_query_profile_indexes",
	"20260716014105_grant_iam_rest_query_runtime_v1",
	"20260730113000_create_iam_rest_query_search_views_v2",
	"20260730120010_create_iam_organization_owner_query_v3",
}

func TestIAMRestQueryMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamRestQueryMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("invalid IAM REST Query migration: %v", err)
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM REST Query migration exceeds 200 lines")
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMRestQueryStorageIsNarrowAndIndexed(t *testing.T) {
	actor := string(readIAMIdentityMigration(t, "fresh", iamRestQueryMigrationNames[0]))
	for _, marker := range []string{
		"SECURITY DEFINER", "SET search_path FROM CURRENT",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()",
		"REVOKE ALL ON FUNCTION iam_rest_query_actor_versions_v1",
	} {
		if !strings.Contains(actor, marker) {
			t.Fatalf("actor version routine lacks %q", marker)
		}
	}
	for _, name := range iamRestQueryMigrationNames[1:4] {
		view := string(readIAMIdentityMigration(t, "fresh", name))
		for _, marker := range []string{
			"SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime'",
			"profile_values",
		} {
			if !strings.Contains(view, marker) {
				t.Fatalf("%s lacks %q", name, marker)
			}
		}
		if strings.Contains(view, "security_barrier") {
			t.Fatalf("%s blocks fixed GIN predicate pushdown", name)
		}
	}
	indexes := string(readIAMIdentityMigration(t, "fresh", iamRestQueryMigrationNames[4]))
	if strings.Count(indexes, "USING gin (profile_values jsonb_path_ops)") != 3 {
		t.Fatal("custom profile storage lacks exact three GIN indexes")
	}
	grant := string(readIAMIdentityMigration(t, "fresh", iamRestQueryMigrationNames[5]))
	if strings.Count(grant, "GRANT SELECT ON iam_rest_query_") != 3 ||
		strings.Count(grant, "GRANT EXECUTE ON FUNCTION") != 1 ||
		strings.Contains(grant, "GRANT SELECT ON iam_users") {
		t.Fatal("IAM REST Query runtime ACL is broader than final views")
	}
	search := string(readIAMIdentityMigration(t, "fresh", iamRestQueryMigrationNames[6]))
	for _, marker := range []string{
		"CREATE VIEW iam_rest_query_users_v2",
		"CREATE VIEW iam_rest_query_roles_v2",
		"CREATE VIEW iam_rest_query_organizations_v2",
		"profile_field.classification = 'public'",
		"AS display_name_folded",
		"AS name_folded",
		"AS description_folded",
		"AS slug_folded",
		"pg_catalog.normalize(pg_catalog.casefold(",
	} {
		if !strings.Contains(search, marker) {
			t.Fatalf("IAM search views lack %q", marker)
		}
	}
	if strings.Count(search, "GRANT SELECT ON iam_rest_query_") != 3 {
		t.Fatal("IAM v2 search view grants are incomplete")
	}
	owners := string(readIAMIdentityMigration(
		t, "fresh", iamRestQueryMigrationNames[7],
	))
	for _, marker := range []string{
		"CREATE VIEW iam_rest_query_organizations_v3",
		"membership.membership_kind = 'owner'",
		"membership.state = 'active'",
		"AS owner_ids",
		"GRANT SELECT ON iam_rest_query_organizations_v3 TO harpy_runtime",
	} {
		if !strings.Contains(owners, marker) {
			t.Fatalf("IAM organization owner view lacks %q", marker)
		}
	}
}
