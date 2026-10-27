-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_password_use_v1(
    p_scope_schema name,
    p_user_id uuid,
    p_expected_user_revision bigint,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_expected_credential_version bigint,
    p_expected_credential_revision bigint,
    p_rehash_legacy boolean,
    p_argon_salt bytea,
    p_argon_tag bytea,
    p_argon_memory_kib integer,
    p_argon_iterations smallint,
    p_argon_parallelism smallint
)
RETURNS TABLE (
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    credential_version bigint,
    credential_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_user_status text;
    v_identity_kind text;
    v_user_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_user_updated_at timestamptz;
    v_credential_state text;
    v_algorithm text;
    v_upgrade_required boolean;
    v_credential_version bigint;
    v_credential_revision bigint;
    v_credential_created_at timestamptz;
    v_credential_updated_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM password use scope is invalid';
    END IF;
    SELECT principal.status, principal.identity_kind, principal.revision,
           principal.token_version, principal.authorization_version,
           principal.updated_at
    INTO v_user_status, v_identity_kind, v_user_revision,
         v_token_version, v_authorization_version, v_user_updated_at
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id
    FOR UPDATE;
    SELECT credential.state, credential.algorithm,
           credential.upgrade_required, credential.credential_version,
           credential.revision, credential.created_at, credential.updated_at
    INTO v_credential_state, v_algorithm, v_upgrade_required,
         v_credential_version, v_credential_revision,
         v_credential_created_at, v_credential_updated_at
    FROM iam_password_credentials AS credential
    WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id
    FOR UPDATE;
    IF v_user_status IS NULL OR v_credential_state IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM password subject is unavailable';
    END IF;
    IF v_user_status IS DISTINCT FROM 'active' OR
       v_identity_kind IS DISTINCT FROM 'human' OR
       v_credential_state IS DISTINCT FROM 'active' OR
       v_user_revision IS DISTINCT FROM p_expected_user_revision OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM
           p_expected_authorization_version OR
       v_credential_version IS DISTINCT FROM
           p_expected_credential_version OR
       v_credential_revision IS DISTINCT FROM p_expected_credential_revision OR
       v_now OPERATOR(pg_catalog.<) v_user_updated_at OR
       v_now OPERATOR(pg_catalog.<) v_credential_updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM password use CAS is stale';
    END IF;
    IF p_rehash_legacy THEN
        IF v_algorithm IS DISTINCT FROM 'bcrypt_v1_adopted' OR
           NOT v_upgrade_required OR
           v_now OPERATOR(pg_catalog.>=)
               v_credential_created_at OPERATOR(pg_catalog.+) INTERVAL '180 days' OR
           pg_catalog.octet_length(p_argon_salt) IS DISTINCT FROM 16 OR
           pg_catalog.octet_length(p_argon_tag) IS DISTINCT FROM 32 OR
           p_argon_memory_kib IS DISTINCT FROM 65536 OR
           p_argon_iterations IS DISTINCT FROM 3::smallint OR
           p_argon_parallelism IS DISTINCT FROM 4::smallint THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM legacy password rehash is invalid';
        END IF;
        UPDATE iam_password_credentials
        SET credential_version =
                v_credential_version OPERATOR(pg_catalog.+) 1,
            algorithm = 'argon2id_v1',
            salt = p_argon_salt,
            tag = p_argon_tag,
            memory_kib = p_argon_memory_kib,
            iterations = p_argon_iterations,
            parallelism = p_argon_parallelism,
            legacy_bcrypt_hash = NULL,
            upgrade_required = false,
            revision = v_credential_revision OPERATOR(pg_catalog.+) 1,
            updated_at = v_now,
            last_used_at = v_now
        WHERE user_id OPERATOR(pg_catalog.=) p_user_id;
        UPDATE iam_users
        SET token_version = v_token_version OPERATOR(pg_catalog.+) 1,
            revision = v_user_revision OPERATOR(pg_catalog.+) 1,
            updated_at = v_now
        WHERE id OPERATOR(pg_catalog.=) p_user_id;
    ELSE
        IF v_algorithm IS DISTINCT FROM 'argon2id_v1' OR
           v_upgrade_required OR p_argon_salt IS NOT NULL OR
           p_argon_tag IS NOT NULL OR p_argon_memory_kib IS NOT NULL OR
           p_argon_iterations IS NOT NULL OR p_argon_parallelism IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM password use profile is invalid';
        END IF;
        UPDATE iam_password_credentials
        SET revision = v_credential_revision OPERATOR(pg_catalog.+) 1,
            updated_at = v_now,
            last_used_at = v_now
        WHERE user_id OPERATOR(pg_catalog.=) p_user_id;
    END IF;
    RETURN QUERY
    SELECT principal.revision, principal.token_version,
           principal.authorization_version, credential.credential_version,
           credential.revision
    FROM iam_users AS principal
    JOIN iam_password_credentials AS credential
      ON credential.user_id OPERATOR(pg_catalog.=) principal.id
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_use_v1(
    name,uuid,bigint,bigint,bigint,bigint,bigint,boolean,
    bytea,bytea,integer,smallint,smallint
) FROM PUBLIC;
