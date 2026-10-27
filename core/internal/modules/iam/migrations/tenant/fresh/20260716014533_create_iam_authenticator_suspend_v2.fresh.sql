-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_authenticator_suspend_v2(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid,
    p_expected_revision bigint, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    SELECT result.result_id, result.subject_user_id, result.result_revision,
        result.token_version, result.authorization_version, result.replayed
    INTO result_id, subject_user_id, result_revision,
        token_version, authorization_version, replayed
    FROM iam_authenticator_suspend_v1(
        p_scope_schema, p_tenant_id, p_operation_id, p_request_sha256,
        p_authenticator_id, p_expected_revision, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS result;
    IF NOT replayed THEN
        UPDATE iam_sessions SET state = 'revoked', revoked_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE user_id = subject_user_id AND state = 'active';
        UPDATE iam_refresh_families SET state = 'revoked',
            revoke_reason = 'authenticator_suspended', revoked_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE user_id = subject_user_id AND state = 'active';
        UPDATE iam_refresh_tokens SET state = 'revoked', revoked_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE family_id IN (
            SELECT id FROM iam_refresh_families WHERE user_id = subject_user_id
        ) AND state = 'active';
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_authenticator_suspend_v2(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_authenticator_suspend_v2(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;

