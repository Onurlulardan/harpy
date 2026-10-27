-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_revoke_all_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_actor_id IS NULL OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.session.revoke_all',
        p_request_sha256, p_actor_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT subject.revision INTO v_user_revision
    FROM iam_users AS subject
    WHERE subject.id = p_actor_id AND subject.status = 'active' FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '23503'; END IF;
    UPDATE iam_sessions SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_actor_id AND state = 'active';
    UPDATE iam_refresh_families SET state = 'revoked',
        revoke_reason = 'all_sessions_revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_actor_id AND state = 'active';
    UPDATE iam_refresh_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE family_id IN (
        SELECT id FROM iam_refresh_families WHERE user_id = p_actor_id
    ) AND state = 'active';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.session.revoke_all', p_request_sha256,
        p_actor_id, p_actor_id, v_user_revision,
        'actor_tenant', p_actor_id, 'session_set', true, true
    ) AS result;
    RETURN QUERY SELECT p_actor_id, p_actor_id, v_user_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_revoke_all_v1(
    name,uuid,uuid,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
