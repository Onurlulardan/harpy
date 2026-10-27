package postgres

import (
	"bytes"
	"strings"
	"testing"
)

func TestCurrentSessionV3PublishesNullableCredentialRevision(t *testing.T) {
	const name = "20260722141000_create_iam_current_session_read_v3"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("current-session v3 migration origin pair differs")
	}
	text := string(fresh)
	for _, marker := range []string{
		"iam_current_session_read_v3(",
		"result_password_credential_revision bigint",
		"LEFT JOIN iam_password_credentials",
		"credential.revision",
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v3",
		"REVOKE EXECUTE ON FUNCTION iam_current_session_read_v2",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("current-session v3 migration lacks %q", marker)
		}
	}
}

func TestCurrentSessionV4PublishesCanonicalIdentity(t *testing.T) {
	const name = "20260727150001_create_iam_current_session_read_v4"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("current-session v4 migration origin pair differs")
	}
	text := string(fresh)
	for _, marker := range []string{
		"iam_current_session_read_v4(",
		"result_email text",
		"result_display_name text",
		"subject.profile_values->>'display_name'",
		"identifier.kind OPERATOR(pg_catalog.=) 'email_v1'",
		"identifier.state OPERATOR(pg_catalog.=) 'active'",
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v4",
		"REVOKE EXECUTE ON FUNCTION iam_current_session_read_v3",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("current-session v4 migration lacks %q", marker)
		}
	}
}

func TestCurrentSessionV4RuntimeACLTargetsBoundOwnerRole(t *testing.T) {
	const name = "20260727150002_repair_iam_current_session_v4_runtime_acl"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("current-session v4 runtime ACL migration origin pair differs")
	}
	text := string(fresh)
	for _, marker := range []string{
		"REVOKE EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)",
		"FROM harpy_v2_runtime",
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)",
		"TO harpy_runtime",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("current-session v4 runtime ACL migration lacks %q", marker)
		}
	}
}

func TestCurrentSessionV5PublishesCanonicalCapabilityOrder(t *testing.T) {
	const name = "20260728090000_create_iam_current_session_read_v5"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("current-session v5 migration origin pair differs")
	}
	text := string(fresh)
	for _, marker := range []string{
		"iam_current_session_read_v5(",
		"pg_catalog.unnest(",
		"session_view.result_tenant_capabilities",
		`ORDER BY capability COLLATE "C"`,
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v5(name,uuid,uuid)",
		"TO harpy_runtime",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("current-session v5 migration lacks %q", marker)
		}
	}
}

func TestCurrentSessionV6ReadsStableProfileKeyAndRetainsV5(t *testing.T) {
	const name = "20260730120020_create_iam_current_session_read_v6"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("current-session v6 migration origin pair differs")
	}
	text := string(fresh)
	for _, marker := range []string{
		"iam_current_session_read_v6(",
		"subject.profile_values -> v_profile_field_id::text",
		"pg_catalog.convert_to(",
		"subject.profile_values::text, 'UTF8'",
		"pg_catalog.jsonb_object_keys(subject.profile_values)",
		"NOT subject.profile_values ? v_profile_field_id::text",
		"subject.profile_values ? 'display_name'",
		"subject.profile_values -> 'display_name'",
		"snapshot.field_key = 'display_name'",
		"snapshot.value_kind = 'string'",
		"snapshot.cardinality = 'single'",
		"snapshot.field_state = 'active'",
		"FROM iam_current_session_read_v5(",
		"session_view.result_email",
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v6(name,uuid,uuid)",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("current-session v6 migration lacks %q", marker)
		}
	}
	if strings.Contains(text, "JOIN iam_profile_fields") {
		t.Fatal("current-session v6 reads mutable profile field registry")
	}
	if strings.Contains(text,
		"REVOKE EXECUTE ON FUNCTION iam_current_session_read_v5") {
		t.Fatal("runtime-24 current-session v5 was revoked")
	}
	if strings.Contains(text, "session_view.result_display_name\n           )") {
		t.Fatal("current-session v6 falls back to unchecked v5 display name")
	}
}
