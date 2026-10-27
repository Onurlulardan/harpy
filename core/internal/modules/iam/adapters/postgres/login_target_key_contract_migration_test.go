package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const loginTargetKeyContractMigration = "20260725124500_repair_iam_login_target_key_contract"

func TestIAMLoginTargetKeyContractMigrationStaysPairedAndBounded(t *testing.T) {
	fresh := readIAMPlatformAuthMigration(t, "fresh", loginTargetKeyContractMigration)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("login target repair exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target repair: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE OR REPLACE FUNCTION platform.iam_resolve_login_target_v1(",
		"NOT platform.iam_hmac_key_id_valid_v1(p_key_id)",
		"pg_catalog.octet_length(p_identifier_hmac) IS DISTINCT FROM 32",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"SECURITY DEFINER", "SET search_path = pg_catalog", "FOR SHARE",
		"REVOKE ALL ON FUNCTION platform.iam_resolve_login_target_v1(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target repair is missing %q", marker)
		}
	}
	if strings.Contains(source, "p_key_id OPERATOR(pg_catalog.!~)") {
		t.Fatal("login target repair retained the obsolete lowercase key-id validator")
	}
}

func TestIAMLoginTargetV3MigrationStaysPairedAndNarrow(t *testing.T) {
	const name = "20260725124610_create_iam_login_target_resolver_v3"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target V3 migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.iam_resolve_login_target_v3(",
		"FROM platform.iam_resolve_login_target_v2(",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v3(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target V3 migration is missing %q", marker)
		}
	}
}

func TestIAMLoginTargetV4MigrationStaysPairedAndCanonical(t *testing.T) {
	const name = "20260725124710_create_iam_login_target_resolver_v4"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target V4 migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.iam_resolve_login_target_v4(",
		"NOT platform.iam_hmac_key_id_valid_v1(p_key_id)",
		"FROM platform.iam_resolve_login_target_v2(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v4(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target V4 migration is missing %q", marker)
		}
	}
}

func TestIAMLoginTargetV5MigrationStaysPairedAndCanonical(t *testing.T) {
	const name = "20260725124810_create_iam_login_target_resolver_v5"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target V5 migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.iam_resolve_login_target_v5(",
		"NOT platform.iam_hmac_key_id_valid_v1(p_key_id)",
		"FROM platform.iam_resolve_login_target_v2(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v5(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target V5 migration is missing %q", marker)
		}
	}
}

func TestIAMLoginTargetV6MigrationStaysPairedAndCanonical(t *testing.T) {
	const name = "20260725124910_create_iam_login_target_resolver_v6"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target V6 migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.iam_resolve_login_target_v6(",
		"NOT platform.iam_hmac_key_id_valid_v1(p_key_id)",
		"FROM platform.iam_resolve_login_target_v2(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v6(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target V6 migration is missing %q", marker)
		}
	}
}

func TestIAMLoginTargetV7MigrationStaysPairedAndCanonical(t *testing.T) {
	const name = "20260725125110_create_iam_login_target_resolver_v7"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect login target V7 migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.iam_resolve_login_target_v7(",
		"NOT platform.iam_hmac_key_id_valid_v1(p_key_id)",
		"FROM platform.iam_resolve_login_target_v2(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v7(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("login target V7 migration is missing %q", marker)
		}
	}
}
