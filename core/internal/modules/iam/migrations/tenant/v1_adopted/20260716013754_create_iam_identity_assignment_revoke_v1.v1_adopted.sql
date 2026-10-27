-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_assignment_revoke_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_assignment_id uuid, p_expected_revision bigint,
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
    v_user_id uuid; v_revision bigint; v_clock bigint; v_user_revision bigint;
BEGIN
    IF p_assignment_id IS NULL OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.assignment.revoke',
        p_request_sha256, p_assignment_id
    );
    IF FOUND THEN RETURN; END IF;
    UPDATE iam_user_role_assignments SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_assignment_id AND revision = p_expected_revision AND state <> 'revoked'
    RETURNING user_id, revision INTO v_user_id, v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM identity assignment revoke compare-and-swap failed'; END IF;
    v_user_revision := iam_identity_user_bump_v1(v_user_id, false);
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.assignment.revoke', p_request_sha256,
        p_assignment_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_assignment_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_assignment_revoke_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
