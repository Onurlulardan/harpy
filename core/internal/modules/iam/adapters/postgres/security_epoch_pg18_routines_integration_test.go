//go:build integration

package postgres

const securityEpochPG18RoutineFixture = `
CREATE FUNCTION iam_platform_session_context_v1(name,uuid,bigint,bigint)
RETURNS TABLE (
    deployment_mode text, tenant_purpose text, subject_id uuid,
    platform_clock bigint
) LANGUAGE sql AS $f$
SELECT 'saas'::text, 'saas_control'::text, $2, 7::bigint
$f$;
CREATE FUNCTION iam_session_issue_v1(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) RETURNS integer LANGUAGE sql AS $f$ SELECT 1 $f$;
CREATE FUNCTION iam_refresh_credential_v2(name,uuid,uuid,bytea,bigint,bytea)
RETURNS TABLE (
    token_id uuid, token_revision bigint, token_state text,
    family_id uuid, family_revision bigint, family_state text,
    session_id uuid, session_revision bigint, session_state text,
    user_id uuid, user_revision bigint, user_status text, identity_kind text,
    token_version bigint, authorization_version bigint,
    access_mode_id uuid, access_mode text, access_mode_state text,
    access_mode_revision bigint, portal_binding_id uuid,
    portal_binding_state text, portal_binding_revision bigint,
    hash_key_id text, secret_hmac bytea, sequence bigint,
    auth_time timestamptz, acr text, amr_password boolean,
    amr_webauthn boolean, amr_totp boolean, amr_recovery boolean,
    step_up_purpose_sha256 bytea, idle_expires_at timestamptz,
    absolute_expires_at timestamptz, retain_until timestamptz
) LANGUAGE plpgsql AS $f$ BEGIN RETURN; END $f$;
CREATE FUNCTION iam_refresh_rotate_v3(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) RETURNS integer LANGUAGE sql AS $f$ SELECT 1 $f$;
CREATE FUNCTION iam_current_session_read_v1(name,uuid,uuid)
RETURNS TABLE (
    result_user_id uuid, result_session_id uuid, result_client_id text,
    result_access_mode text, result_portal_binding_id uuid,
    result_token_version bigint, result_authorization_version bigint,
    result_permissions text[]
) LANGUAGE sql AS $f$
SELECT $2, $3, 'harpy-operator-console-v1'::text,
       'operator_console'::text, NULL::uuid, 1::bigint, 1::bigint,
       ARRAY['billing.read']::text[]
$f$;`
