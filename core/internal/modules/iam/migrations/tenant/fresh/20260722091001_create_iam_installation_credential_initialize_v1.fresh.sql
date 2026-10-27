-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_credential_initialize_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_tenant_id uuid,
    p_user_id uuid,
    p_credential_id uuid,
    p_credential_version bigint,
    p_salt bytea,
    p_tag bytea
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    user_id uuid,
    credential_id uuid,
    credential_version bigint,
    credential_revision bigint,
    marker_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_existing iam_installation_password_credentials%ROWTYPE;
    v_user_status text;
    v_identity_kind text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation credential scope is invalid';
    END IF;
    IF p_operation_id IS NULL OR p_tenant_id IS NULL OR p_user_id IS NULL OR
       p_credential_id IS NULL OR
       p_credential_version IS DISTINCT FROM 1::bigint OR
       pg_catalog.octet_length(p_salt) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_tag) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM installation credential evidence is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM iam_installation_identity_facts AS identity
        WHERE identity.operation_id OPERATOR(pg_catalog.=) p_operation_id AND
              identity.tenant_id OPERATOR(pg_catalog.=) p_tenant_id AND
              identity.first_user_id OPERATOR(pg_catalog.=) p_user_id
        FOR UPDATE
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation credential plan changed';
    END IF;
    SELECT marker.*
    INTO v_existing
    FROM iam_installation_password_credentials AS marker
    WHERE marker.mutation_operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_existing.user_id IS DISTINCT FROM p_user_id OR
           v_existing.credential_id IS DISTINCT FROM p_credential_id OR
           v_existing.credential_version IS DISTINCT FROM p_credential_version OR
           v_existing.state IS DISTINCT FROM 'active' OR NOT EXISTS (
                SELECT 1 FROM iam_password_credentials AS credential
                WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id AND
                      credential.credential_version OPERATOR(pg_catalog.=)
                          p_credential_version AND
                      credential.state OPERATOR(pg_catalog.=) 'reset_required' AND
                      credential.salt OPERATOR(pg_catalog.=) p_salt AND
                      credential.tag OPERATOR(pg_catalog.=) p_tag
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM installation credential replay conflicts';
        END IF;
        RETURN QUERY SELECT p_operation_id, p_tenant_id, p_user_id,
            p_credential_id, p_credential_version, 1::bigint,
            v_existing.revision;
        RETURN;
    END IF;
    SELECT principal.status, principal.identity_kind
    INTO v_user_status, v_identity_kind
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id
    FOR UPDATE;
    IF v_user_status IS DISTINCT FROM 'active' OR
       v_identity_kind IS DISTINCT FROM 'human' OR EXISTS (
            SELECT 1 FROM iam_password_credentials AS credential
            WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation credential subject changed';
    END IF;
    INSERT INTO iam_password_credentials (
        user_id, credential_version, state, algorithm, salt, tag,
        memory_kib, iterations, parallelism, legacy_bcrypt_hash,
        upgrade_required, revision, created_at, updated_at,
        last_used_at, invalidated_at
    ) VALUES (
        p_user_id, p_credential_version, 'reset_required', 'argon2id_v1',
        p_salt, p_tag, 65536, 3, 4, NULL, false, 1, v_now, v_now, NULL, NULL
    );
    INSERT INTO iam_installation_password_credentials (
        credential_id, user_id, installation_operation_id,
        mutation_operation_id, credential_version, purpose, state,
        superseded_by, revision, created_at, updated_at, retired_at
    ) VALUES (
        p_credential_id, p_user_id, p_operation_id, p_operation_id,
        p_credential_version, 'installation_initial_v1', 'active',
        NULL, 1, v_now, v_now, NULL
    );
    RETURN QUERY SELECT p_operation_id, p_tenant_id, p_user_id,
        p_credential_id, p_credential_version, 1::bigint, 1::bigint;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_credential_initialize_v1(
    name,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_credential_initialize_v1(
    name,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) TO harpy_runtime;
