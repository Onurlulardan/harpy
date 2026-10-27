//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamSessionV6Schema = "t_session_v6"

func TestIAMCurrentSessionV6ProfileFallbackAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "session_v6_"+origin)
			prepareIAMSessionV6PG18(t, ctx, database, origin)
			assertIAMSessionV6ProfileFallbacks(t, ctx, database)
		})
	}
}

func prepareIAMSessionV6PG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE SCHEMA `+iamSessionV6Schema+`;
GRANT USAGE ON SCHEMA `+iamSessionV6Schema+` TO harpy_runtime;`); err != nil {
		t.Fatal(err)
	}
	applyIAMTenantDocuments(t, ctx, database, iamSessionV6Schema,
		[]iamFoundationDocument{{
			path: "session-v6-fixture",
			content: []byte(`
CREATE TABLE iam_users (
 id uuid PRIMARY KEY, status text NOT NULL,
 profile_schema_version_id uuid NOT NULL, profile_values jsonb NOT NULL,
 profile_sha256 bytea NOT NULL
);
CREATE TABLE iam_profile_field_revisions (
 schema_version_id uuid NOT NULL, resource_kind text NOT NULL,
 field_id uuid NOT NULL, field_key text NOT NULL, value_kind text NOT NULL,
 cardinality text NOT NULL, field_state text NOT NULL
);
CREATE FUNCTION iam_current_session_read_v5(
 p_scope_schema name,p_user_id uuid,p_session_id uuid
) RETURNS TABLE (
 result_user_id uuid,result_session_id uuid,result_client_id text,
 result_access_mode text,result_portal_binding_id uuid,
 result_token_version bigint,result_authorization_version bigint,
 result_deployment_mode text,result_tenant_purpose text,
 result_platform_subject_id uuid,result_platform_clock bigint,
 result_security_epoch bytea,result_tenant_capabilities text[],
 result_password_credential_revision bigint,result_email text,
 result_display_name text
) LANGUAGE sql AS $fixture$
 SELECT p_user_id,p_session_id,'browser','tenant',NULL::uuid,1::bigint,0::bigint,
  'local','primary',NULL::uuid,0::bigint,pg_catalog.decode(
   pg_catalog.repeat('00',32),'hex'),ARRAY[]::text[],1::bigint,
  p_user_id::text||'@example.test','unchecked-v5'
$fixture$;`),
		}, {
			path: "20260730120020_create_iam_current_session_read_v6",
			content: readIAMIdentityMigration(t, origin,
				"20260730120020_create_iam_current_session_read_v6"),
		}})
	seed := `
INSERT INTO iam_profile_field_revisions (
 schema_version_id,resource_kind,field_id,field_key,value_kind,
 cardinality,field_state
) VALUES
 ('72000000-0000-4000-8000-000000000001','user',
  '73000000-0000-7000-8000-000000000001','display_name',
  'string','single','active'),
 ('72000000-0000-4000-8000-000000000002','user',
  '73000000-0000-7000-8000-000000000002','display_name',
  'string','single','active'),
 ('72000000-0000-4000-8000-000000000003','user',
  '73000000-0000-7000-8000-000000000003','display_name',
  'string','single','active'),
 ('72000000-0000-4000-8000-000000000004','user',
  '73000000-0000-7000-8000-000000000004','display_name',
  'string','single','active');
INSERT INTO iam_users (
 id,status,profile_schema_version_id,profile_values,profile_sha256
) VALUES
 ('74000000-0000-4000-8000-000000000001','active',
  '72000000-0000-4000-8000-000000000001',
  '{"73000000-0000-7000-8000-000000000001":"Local Instance Owner"}',
  pg_catalog.sha256(pg_catalog.convert_to(
   '{"73000000-0000-7000-8000-000000000001":"Local Instance Owner"}','UTF8'))),
 ('74000000-0000-4000-8000-000000000002','active',
  '72000000-0000-4000-8000-000000000002','{"display_name":"Legacy Owner"}',
  pg_catalog.sha256(pg_catalog.convert_to(
   '{"display_name":"Legacy Owner"}'::jsonb::text,'UTF8'))),
 ('74000000-0000-4000-8000-000000000003','active',
  '72000000-0000-4000-8000-000000000003',
  '{"display_name":"Spoofed","extra":"value"}',
  pg_catalog.sha256(pg_catalog.convert_to(
   '{"display_name":"Spoofed","extra":"value"}'::jsonb::text,'UTF8'))),
 ('74000000-0000-4000-8000-000000000004','active',
  '72000000-0000-4000-8000-000000000004','{"display_name":"Tampered"}',
  pg_catalog.decode(pg_catalog.repeat('ff',32),'hex'));`
	mustExecIAMTenant(t, ctx, database, iamSessionV6Schema, seed)
}

func assertIAMSessionV6ProfileFallbacks(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	cases := []struct{ user, want string }{
		{"74000000-0000-4000-8000-000000000001", "Local Instance Owner"},
		{"74000000-0000-4000-8000-000000000002", "Legacy Owner"},
		{"74000000-0000-4000-8000-000000000003",
			"74000000-0000-4000-8000-000000000003@example.test"},
		{"74000000-0000-4000-8000-000000000004",
			"74000000-0000-4000-8000-000000000004@example.test"},
	}
	for _, fixture := range cases {
		got, err := readIAMSessionV6DisplayName(
			ctx, database, fixture.user, "75000000-0000-4000-8000-000000000001",
		)
		if err != nil || got != fixture.want {
			t.Fatalf("session v6 display name for %s = %q, want %q: %v",
				fixture.user, got, fixture.want, err)
		}
	}
	mustExecIAMTenant(t, ctx, database, iamSessionV6Schema, `
INSERT INTO iam_profile_field_revisions
SELECT schema_version_id,resource_kind,
 '73000000-0000-7000-8000-000000000099','display_name',
 value_kind,cardinality,field_state
FROM iam_profile_field_revisions
WHERE schema_version_id='72000000-0000-4000-8000-000000000004';`)
	_, err := readIAMSessionV6DisplayName(
		ctx, database, "74000000-0000-4000-8000-000000000004",
		"75000000-0000-4000-8000-000000000001",
	)
	assertIAMSQLState(t, err, "23514")
}

func readIAMSessionV6DisplayName(
	ctx context.Context, database *sql.DB, userID, sessionID string,
) (string, error) {
	var displayName string
	err := database.QueryRowContext(ctx, `
SELECT result_display_name FROM `+iamSessionV6Schema+`.iam_current_session_read_v6(
 $1::name,$2::uuid,$3::uuid)`,
		iamSessionV6Schema, userID, sessionID).Scan(&displayName)
	return displayName, err
}
