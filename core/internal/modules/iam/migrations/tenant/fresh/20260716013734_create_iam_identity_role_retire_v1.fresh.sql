-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_role_retire_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_role_id uuid, p_expected_revision bigint,
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
    IF p_role_id IS NULL OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) OR EXISTS (
          SELECT 1 FROM iam_user_role_assignments
          WHERE role_id = p_role_id AND state <> 'revoked'
       ) OR EXISTS (
          SELECT 1 FROM iam_role_permission_bindings
          WHERE role_id = p_role_id AND state = 'active'
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity role retirement has dependencies'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.role.retire', p_request_sha256, p_role_id
    );
    IF FOUND THEN RETURN; END IF;
    UPDATE iam_roles SET status = 'retired', retired_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_role_id AND revision = p_expected_revision AND status = 'active'
      AND NOT is_protected AND NOT is_default RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM identity role retire compare-and-swap failed'; END IF;
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.role.retire', p_request_sha256,
        p_role_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_role_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_role_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
