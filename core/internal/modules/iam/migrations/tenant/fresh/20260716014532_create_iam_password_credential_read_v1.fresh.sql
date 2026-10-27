-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_password_credential_read_v1(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    algorithm text, salt bytea, tag bytea, memory_kib integer,
    iterations smallint, parallelism smallint, legacy_bcrypt_hash text,
    credential_version bigint, credential_revision bigint, created_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_user_id IS NULL OR p_user_id <> p_actor_id OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT credential.algorithm, credential.salt, credential.tag,
        credential.memory_kib, credential.iterations, credential.parallelism,
        credential.legacy_bcrypt_hash, credential.credential_version,
        credential.revision, credential.created_at
    FROM iam_password_credentials AS credential
    WHERE credential.user_id = p_user_id AND credential.state = 'active';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_credential_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_password_credential_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;

