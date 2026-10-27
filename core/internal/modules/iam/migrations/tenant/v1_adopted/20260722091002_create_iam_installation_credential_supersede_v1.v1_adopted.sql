-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_credential_supersede_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_tenant_id uuid,
    p_user_id uuid,
    p_current_credential_id uuid,
    p_current_version bigint,
    p_next_credential_id uuid,
    p_next_version bigint,
    p_salt bytea,
    p_tag bytea
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    user_id uuid,
    previous_credential_id uuid,
    credential_id uuid,
    credential_version bigint,
    credential_revision bigint,
    previous_marker_revision bigint,
    marker_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_previous iam_installation_password_credentials%ROWTYPE;
    v_replay iam_installation_password_credentials%ROWTYPE;
    v_credential_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM credential supersede scope is invalid';
    END IF;
    IF p_operation_id IS NULL OR p_tenant_id IS NULL OR p_user_id IS NULL OR
       p_current_credential_id IS NULL OR p_next_credential_id IS NULL OR
       p_current_credential_id OPERATOR(pg_catalog.=) p_next_credential_id OR
       p_current_version OPERATOR(pg_catalog.<) 1 OR
       p_next_version IS DISTINCT FROM
           p_current_version OPERATOR(pg_catalog.+) 1 OR
       pg_catalog.octet_length(p_salt) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_tag) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM credential supersede evidence is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_installation_identity_facts AS identity
        WHERE identity.tenant_id OPERATOR(pg_catalog.=) p_tenant_id AND
              identity.first_user_id OPERATOR(pg_catalog.=) p_user_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM credential supersede tenant is invalid';
    END IF;
    SELECT marker.*
    INTO v_replay
    FROM iam_installation_password_credentials AS marker
    WHERE marker.mutation_operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_replay.user_id IS DISTINCT FROM p_user_id OR
           v_replay.credential_id IS DISTINCT FROM p_next_credential_id OR
           v_replay.credential_version IS DISTINCT FROM p_next_version OR
           v_replay.state IS DISTINCT FROM 'active' OR NOT EXISTS (
                SELECT 1 FROM iam_password_credentials AS credential
                WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id AND
                      credential.credential_version OPERATOR(pg_catalog.=)
                          p_next_version AND
                      credential.state OPERATOR(pg_catalog.=) 'reset_required' AND
                      credential.salt OPERATOR(pg_catalog.=) p_salt AND
                      credential.tag OPERATOR(pg_catalog.=) p_tag
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM credential supersede replay conflicts';
        END IF;
        SELECT marker.*
        INTO STRICT v_previous
        FROM iam_installation_password_credentials AS marker
        WHERE marker.superseded_by OPERATOR(pg_catalog.=) p_next_credential_id;
        SELECT credential.revision
        INTO STRICT v_credential_revision
        FROM iam_password_credentials AS credential
        WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id;
        RETURN QUERY SELECT p_operation_id, p_tenant_id, p_user_id,
            p_current_credential_id, p_next_credential_id, p_next_version,
            v_credential_revision, v_previous.revision, v_replay.revision;
        RETURN;
    END IF;
    SELECT marker.*
    INTO v_previous
    FROM iam_installation_password_credentials AS marker
    WHERE marker.credential_id OPERATOR(pg_catalog.=) p_current_credential_id
    FOR UPDATE;
    IF NOT FOUND OR v_previous.user_id IS DISTINCT FROM p_user_id OR
       v_previous.credential_version IS DISTINCT FROM p_current_version OR
       v_previous.state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM credential supersede predecessor changed';
    END IF;
    UPDATE iam_password_credentials AS credential
    SET credential_version = p_next_version, state = 'reset_required',
        algorithm = 'argon2id_v1', salt = p_salt, tag = p_tag,
        memory_kib = 65536, iterations = 3, parallelism = 4,
        legacy_bcrypt_hash = NULL, upgrade_required = false,
        revision = credential.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, last_used_at = NULL, invalidated_at = NULL
    WHERE credential.user_id OPERATOR(pg_catalog.=) p_user_id AND
          credential.credential_version OPERATOR(pg_catalog.=)
              p_current_version AND
          credential.state OPERATOR(pg_catalog.=) 'reset_required'
    RETURNING credential.revision INTO v_credential_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM credential supersede CAS changed';
    END IF;
    INSERT INTO iam_installation_password_credentials (
        credential_id, user_id, installation_operation_id,
        mutation_operation_id, credential_version, purpose, state,
        superseded_by, revision, created_at, updated_at, retired_at
    ) VALUES (
        p_next_credential_id, p_user_id,
        v_previous.installation_operation_id, p_operation_id,
        p_next_version, 'installation_initial_v1', 'active',
        NULL, 1, v_now, v_now, NULL
    );
    UPDATE iam_installation_password_credentials AS marker
    SET state = 'superseded', superseded_by = p_next_credential_id,
        revision = marker.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now, retired_at = v_now
    WHERE marker.credential_id OPERATOR(pg_catalog.=)
        p_current_credential_id;
    RETURN QUERY SELECT p_operation_id, p_tenant_id, p_user_id,
        p_current_credential_id, p_next_credential_id, p_next_version,
        v_credential_revision,
        v_previous.revision OPERATOR(pg_catalog.+) 1, 1::bigint;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_credential_supersede_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,bigint,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_credential_supersede_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,bigint,bytea,bytea
) TO harpy_runtime;
