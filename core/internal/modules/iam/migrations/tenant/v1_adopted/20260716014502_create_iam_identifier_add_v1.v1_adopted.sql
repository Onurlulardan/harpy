-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identifier_add_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_identifier_id uuid, p_user_id uuid,
    p_kind text, p_display_value text, p_normalized_value text,
    p_normalized_sha256 bytea, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_identifier_id IS NULL OR p_user_id IS NULL OR
       p_kind NOT IN ('email_v1', 'portal_v1') OR
       p_display_value IS NULL OR p_normalized_value IS NULL OR
       pg_catalog.octet_length(p_normalized_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_normalized_value, 'UTF8'))
           IS DISTINCT FROM p_normalized_sha256 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM identifier add input is invalid';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.identifier.add', p_request_sha256, p_identifier_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_users AS subject
        WHERE subject.id = p_user_id AND subject.status IN ('pending', 'active')
        FOR UPDATE
    ) THEN RAISE EXCEPTION USING ERRCODE = '23503'; END IF;
    INSERT INTO iam_user_identifiers (
        id, user_id, kind, display_value, normalized_value,
        normalized_sha256, state, revision, created_at, updated_at
    ) VALUES (
        p_identifier_id, p_user_id, p_kind, p_display_value,
        p_normalized_value, p_normalized_sha256, 'pending', 1, v_now, v_now
    );
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.identifier.add', p_request_sha256,
        p_identifier_id, p_user_id, 1, 'actor_tenant', p_actor_id,
        'identifier', true, true
    ) AS result;
    RETURN QUERY SELECT p_identifier_id, p_user_id, 1::bigint,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identifier_add_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,text,text,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
