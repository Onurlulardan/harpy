-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_update_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_user_id uuid, p_expected_revision bigint,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
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
    v_values jsonb; v_revision bigint; v_clock bigint;
BEGIN
    IF p_user_id IS NULL OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.user.update', p_request_sha256, p_user_id
    );
    IF FOUND THEN RETURN; END IF;
    v_values := iam_identity_profile_assert_v1(
        'user', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    UPDATE iam_users SET profile_schema_version_id = p_profile_schema_id,
        profile_values = v_values, profile_sha256 = p_profile_sha256,
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_user_id AND revision = p_expected_revision
      AND status <> 'tombstoned' RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM identity user update compare-and-swap failed';
    END IF;
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.user.update', p_request_sha256,
        p_user_id, v_revision, p_profile_sha256
    );
    RETURN QUERY SELECT p_user_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_update_v1(
    name,uuid,bytea,uuid,bigint,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
