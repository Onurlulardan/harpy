-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_revoke_one_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_session_id uuid, p_expected_revision bigint,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_session_id IS NULL OR p_expected_revision < 1 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.session.revoke_one',
        p_request_sha256, p_session_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    UPDATE iam_sessions AS session SET state = 'revoked', revoked_at = v_now,
        revision = session.revision + 1, updated_at = v_now
    WHERE session.id = p_session_id AND session.user_id = p_actor_id
      AND session.state = 'active' AND session.revision = p_expected_revision
    RETURNING session.user_id, session.revision INTO v_user_id, v_revision;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    UPDATE iam_refresh_families SET state = 'revoked',
        revoke_reason = 'session_revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE session_id = p_session_id AND state = 'active';
    UPDATE iam_refresh_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE family_id IN (
        SELECT id FROM iam_refresh_families WHERE session_id = p_session_id
    ) AND state = 'active';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.session.revoke_one', p_request_sha256,
        p_session_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'session', false, false
    ) AS result;
    RETURN QUERY SELECT p_session_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_revoke_one_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
