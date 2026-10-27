-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_password_enroll_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_user_id uuid, p_salt bytea, p_tag bytea,
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
    v_clock bigint;
BEGIN
    IF p_operation_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_user_id IS NULL OR
       pg_catalog.octet_length(p_salt) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_tag) IS DISTINCT FROM 32 OR
       NOT EXISTS (
          SELECT 1 FROM iam_users AS subject
          WHERE subject.id = p_user_id AND subject.status = 'active'
       ) OR NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501',
          MESSAGE = 'IAM password enrollment input or authority is invalid';
    END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.user.password_enroll',
        p_request_sha256, p_user_id
    );
    IF FOUND THEN RETURN; END IF;
    INSERT INTO iam_password_credentials (
        user_id, credential_version, state, algorithm, salt, tag,
        memory_kib, iterations, parallelism, legacy_bcrypt_hash,
        upgrade_required, revision, created_at, updated_at
    ) VALUES (
        p_user_id, 1, 'active', 'argon2id_v1', p_salt, p_tag,
        65536, 3, 4, NULL, false, 1, v_now, v_now
    );
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.user.password_enroll',
        p_request_sha256, p_user_id, 1,
        pg_catalog.sha256(p_salt || p_tag)
    );
    RETURN QUERY SELECT p_user_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_password_enroll_v1(
    name,uuid,bytea,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_identity_password_enroll_v1(
    name,uuid,bytea,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
