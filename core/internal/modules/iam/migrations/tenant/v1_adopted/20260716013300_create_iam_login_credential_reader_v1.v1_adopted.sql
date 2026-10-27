-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_login_credential_v1(
    p_scope_schema name,
    p_source_user_id uuid,
    p_source_identifier_id uuid,
    p_identifier_kind text,
    p_source_revision bigint,
    p_source_sha256 bytea,
    p_access_mode text
)
RETURNS TABLE (
    identifier_id uuid,
    user_id uuid,
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    access_mode_id uuid,
    access_mode_revision bigint,
    portal_binding_id uuid,
    portal_binding_revision bigint,
    credential_version bigint,
    credential_revision bigint,
    algorithm text,
    salt bytea,
    tag bytea,
    memory_kib integer,
    iterations smallint,
    parallelism smallint,
    legacy_bcrypt_hash text,
    credential_created_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_rows bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM login credential scope is invalid';
    END IF;
    IF p_source_user_id IS NULL OR p_source_identifier_id IS NULL OR
       p_source_revision OPERATOR(pg_catalog.<=) 0 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       NOT (
           (p_identifier_kind IS NOT DISTINCT FROM 'email_v1' AND
            p_access_mode IS NOT DISTINCT FROM 'operator_console') OR
           (p_identifier_kind IS NOT DISTINCT FROM 'portal_v1' AND
            p_access_mode IS NOT DISTINCT FROM 'portal')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM login credential evidence is invalid';
    END IF;
    RETURN QUERY
    SELECT identifier.id, principal.id, principal.revision,
           principal.token_version, principal.authorization_version,
           access.id, access.revision, binding.id, binding.revision,
           credential.credential_version, credential.revision,
           credential.algorithm, credential.salt, credential.tag,
           credential.memory_kib, credential.iterations,
           credential.parallelism, credential.legacy_bcrypt_hash,
           credential.created_at
    FROM iam_user_identifiers AS identifier
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) identifier.user_id
    JOIN iam_user_access_modes AS access
      ON access.user_id OPERATOR(pg_catalog.=) principal.id
     AND access.mode OPERATOR(pg_catalog.=) p_access_mode
    JOIN iam_password_credentials AS credential
      ON credential.user_id OPERATOR(pg_catalog.=) principal.id
    LEFT JOIN iam_portal_subject_bindings AS binding
      ON binding.user_id OPERATOR(pg_catalog.=) principal.id
     AND binding.access_mode_id OPERATOR(pg_catalog.=) access.id
     AND binding.state OPERATOR(pg_catalog.=) 'active'
    WHERE identifier.kind OPERATOR(pg_catalog.=) p_identifier_kind
      AND identifier.id OPERATOR(pg_catalog.=) p_source_identifier_id
      AND identifier.user_id OPERATOR(pg_catalog.=) p_source_user_id
      AND principal.id OPERATOR(pg_catalog.=) p_source_user_id
      AND identifier.revision OPERATOR(pg_catalog.=) p_source_revision
      AND identifier.normalized_sha256 OPERATOR(pg_catalog.=) p_source_sha256
      AND identifier.state OPERATOR(pg_catalog.=) 'active'
      AND principal.status OPERATOR(pg_catalog.=) 'active'
      AND principal.identity_kind OPERATOR(pg_catalog.=) 'human'
      AND access.state OPERATOR(pg_catalog.=) 'active'
      AND credential.state OPERATOR(pg_catalog.=) 'active'
      AND (
          (p_access_mode OPERATOR(pg_catalog.=) 'operator_console' AND
           binding.id IS NULL) OR
          (p_access_mode OPERATOR(pg_catalog.=) 'portal' AND
           binding.id IS NOT NULL)
      );
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows OPERATOR(pg_catalog.>) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '21000',
            MESSAGE = 'IAM login credential evidence is ambiguous';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_login_credential_v1(
    name,uuid,uuid,text,bigint,bytea,text
) FROM PUBLIC;
