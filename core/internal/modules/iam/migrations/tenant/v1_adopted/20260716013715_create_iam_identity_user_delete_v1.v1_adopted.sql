-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_delete_v1(
    p_scope_schema name, p_operation_id uuid,
    p_request_sha256 bytea, p_user_id uuid, p_expected_revision bigint,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_revision bigint; v_clock bigint;
BEGIN
    IF p_user_id IS NULL OR p_user_id = p_actor_id OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) OR EXISTS (
          SELECT 1 FROM iam_organization_memberships
          WHERE user_id = p_user_id AND state <> 'tombstoned'
       ) OR EXISTS (
          SELECT 1 FROM iam_user_role_assignments
          WHERE user_id = p_user_id AND state <> 'revoked'
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity user deletion is denied'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.user.delete', p_request_sha256, p_user_id
    );
    IF FOUND THEN RETURN; END IF;
    UPDATE iam_users SET status = 'tombstoned', token_version = token_version + 1,
        authorization_version = authorization_version + 1,
        tombstoned_at = v_now, revision = revision + 1, updated_at = v_now
    WHERE id = p_user_id AND revision = p_expected_revision AND status = 'disabled'
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM identity user delete compare-and-swap failed';
    END IF;
    UPDATE iam_user_identifiers SET state = 'tombstoned', tombstoned_at = v_now,
        cooldown_until = v_now + INTERVAL '30 days', revision = revision + 1,
        updated_at = v_now WHERE user_id = p_user_id AND state = 'active';
    UPDATE iam_user_access_modes SET state = 'retired', retired_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_user_id AND state <> 'retired';
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.user.delete', p_request_sha256,
        p_user_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_user_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_delete_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
