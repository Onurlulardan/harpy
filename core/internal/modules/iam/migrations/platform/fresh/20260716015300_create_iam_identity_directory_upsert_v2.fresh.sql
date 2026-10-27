-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_identity_directory_upsert_v2(
    p_tenant_id uuid, p_group_generation bigint, p_iam_generation bigint,
    p_user_id uuid, p_identifier_id uuid, p_lookup_key_id text,
    p_identifier_hmac bytea, p_source_revision bigint, p_source_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_row platform.iam_login_directory%ROWTYPE;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_user_id IS NULL OR p_identifier_id IS NULL OR
       NOT platform.iam_hmac_key_id_valid_v1(p_lookup_key_id) OR
       p_identifier_hmac IS NULL OR pg_catalog.octet_length(p_identifier_hmac) <> 32 OR
       p_source_revision IS NULL OR p_source_revision < 1 OR
       p_source_sha256 IS NULL OR pg_catalog.octet_length(p_source_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity directory upsert input is invalid';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, p_group_generation, p_iam_generation
    );
    SELECT directory.* INTO v_row
    FROM platform.iam_login_directory AS directory
    WHERE directory.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
      AND directory.identifier_kind OPERATOR(pg_catalog.=) 'email_v1'
      AND directory.lookup_key_id OPERATOR(pg_catalog.=) p_lookup_key_id
      AND directory.identifier_hmac OPERATOR(pg_catalog.=) p_identifier_hmac
    FOR UPDATE;
    IF FOUND THEN
        IF v_row.user_id IS DISTINCT FROM p_user_id OR
           v_row.source_identifier_id IS DISTINCT FROM p_identifier_id OR
           v_row.source_revision OPERATOR(pg_catalog.>) p_source_revision THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM identity directory ownership or revision conflicts';
        END IF;
        UPDATE platform.iam_login_directory AS directory
        SET source_revision = p_source_revision, state = 'active',
            source_sha256 = p_source_sha256, projected_at = v_now,
            tombstoned_at = NULL
        WHERE directory.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
          AND directory.identifier_kind OPERATOR(pg_catalog.=) 'email_v1'
          AND directory.lookup_key_id OPERATOR(pg_catalog.=) p_lookup_key_id
          AND directory.identifier_hmac OPERATOR(pg_catalog.=) p_identifier_hmac;
    ELSE
        INSERT INTO platform.iam_login_directory (
            tenant_id, identifier_kind, lookup_key_id, identifier_hmac,
            user_id, source_identifier_id, source_revision, state,
            source_sha256, projected_at
        ) VALUES (
            p_tenant_id, 'email_v1', p_lookup_key_id, p_identifier_hmac,
            p_user_id, p_identifier_id, p_source_revision, 'active',
            p_source_sha256, v_now
        );
    END IF;
    RETURN p_source_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_identity_directory_upsert_v2(
    uuid,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_upsert_v2(
    uuid,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea
) TO forgecrud_runtime;
