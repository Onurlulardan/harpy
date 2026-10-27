-- +goose Up
-- Migration stream version: 20260730120020.
-- +goose StatementBegin
CREATE FUNCTION iam_current_session_read_v6(
    p_scope_schema name, p_user_id uuid, p_session_id uuid
)
RETURNS TABLE (
    result_user_id uuid, result_session_id uuid, result_client_id text,
    result_access_mode text, result_portal_binding_id uuid,
    result_token_version bigint, result_authorization_version bigint,
    result_deployment_mode text, result_tenant_purpose text,
    result_platform_subject_id uuid, result_platform_clock bigint,
    result_security_epoch bytea, result_tenant_capabilities text[],
    result_password_credential_revision bigint, result_email text,
    result_display_name text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_profile_field_id uuid;
    v_display_count bigint;
    v_exact_count bigint;
BEGIN
    SELECT
        pg_catalog.count(*) FILTER (
            WHERE snapshot.field_key = 'display_name'
        ),
        pg_catalog.count(*) FILTER (
            WHERE snapshot.field_key = 'display_name'
              AND snapshot.value_kind = 'string'
              AND snapshot.cardinality = 'single'
              AND snapshot.field_state = 'active'
        )
    INTO v_display_count, v_exact_count
    FROM iam_users AS subject
    LEFT JOIN iam_profile_field_revisions AS snapshot
      ON snapshot.schema_version_id = subject.profile_schema_version_id
     AND snapshot.resource_kind = 'user'
    WHERE subject.id = p_user_id;
    IF v_display_count IS DISTINCT FROM 1
       OR v_exact_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM current session profile snapshot is invalid';
    END IF;
    SELECT snapshot.field_id
    INTO STRICT v_profile_field_id
    FROM iam_users AS subject
    JOIN iam_profile_field_revisions AS snapshot
      ON snapshot.schema_version_id = subject.profile_schema_version_id
     AND snapshot.resource_kind = 'user'
    WHERE subject.id = p_user_id
      AND snapshot.field_key = 'display_name'
      AND snapshot.value_kind = 'string'
      AND snapshot.cardinality = 'single'
      AND snapshot.field_state = 'active';
    RETURN QUERY
    SELECT session_view.result_user_id, session_view.result_session_id,
           session_view.result_client_id, session_view.result_access_mode,
           session_view.result_portal_binding_id,
           session_view.result_token_version,
           session_view.result_authorization_version,
           session_view.result_deployment_mode,
           session_view.result_tenant_purpose,
           session_view.result_platform_subject_id,
           session_view.result_platform_clock,
           session_view.result_security_epoch,
           session_view.result_tenant_capabilities,
           session_view.result_password_credential_revision,
           session_view.result_email,
           COALESCE(
               CASE WHEN pg_catalog.jsonb_typeof(
                        subject.profile_values -> v_profile_field_id::text
                    ) = 'string' THEN NULLIF(pg_catalog.btrim(
                   subject.profile_values ->> v_profile_field_id::text
               ), '') END,
               CASE WHEN subject.profile_sha256 =
                       pg_catalog.sha256(pg_catalog.convert_to(
                           subject.profile_values::text, 'UTF8'
                       ))
                    AND pg_catalog.jsonb_typeof(subject.profile_values) = 'object'
                    AND (SELECT pg_catalog.count(*) FROM
                        pg_catalog.jsonb_object_keys(subject.profile_values)) = 1
                    AND NOT subject.profile_values ? v_profile_field_id::text
                    AND subject.profile_values ? 'display_name'
                    AND pg_catalog.jsonb_typeof(
                        subject.profile_values -> 'display_name'
                    ) = 'string'
                   THEN NULLIF(pg_catalog.btrim(
                       subject.profile_values ->> 'display_name'
                   ), '') END,
               session_view.result_email
           )
    FROM iam_current_session_read_v5(p_scope_schema, p_user_id, p_session_id)
         AS session_view
    JOIN iam_users AS subject
      ON subject.id OPERATOR(pg_catalog.=) session_view.result_user_id
     AND subject.status OPERATOR(pg_catalog.=) 'active';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_current_session_read_v6(name,uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v6(name,uuid,uuid)
    TO forgecrud_runtime;
