-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_login_target_v1(
    p_tenant_id uuid,
    p_identifier_kind text,
    p_key_id text,
    p_identifier_hmac bytea
)
RETURNS TABLE (
    tenant_id uuid,
    source_revision bigint,
    source_sha256 bytea,
    group_generation bigint,
    iam_generation bigint,
    layout_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR p_identifier_kind IS NULL OR NOT (
           p_identifier_kind OPERATOR(pg_catalog.=) ANY (
               ARRAY['email_v1','portal_v1']::text[]
           )
       ) OR p_key_id IS NULL OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       p_key_id OPERATOR(pg_catalog.!~) '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       p_identifier_hmac IS NULL OR
       pg_catalog.octet_length(p_identifier_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM login target input is invalid';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'harpy:iam-login-target:' || p_tenant_id::text || ':' ||
        p_identifier_kind || ':' || p_key_id || ':' ||
        pg_catalog.encode(p_identifier_hmac, 'hex'), 0
    ));
    SELECT target.group_generation, target.iam_generation,
           target.layout_generation
    INTO group_generation, iam_generation, layout_generation
    FROM platform.iam_resolve_tenant_target_v1(p_tenant_id) AS target;

    BEGIN
        SELECT directory.tenant_id, directory.source_revision,
               directory.source_sha256
        INTO STRICT tenant_id, source_revision, source_sha256
        FROM platform.iam_login_directory AS directory
        WHERE directory.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
          AND directory.identifier_kind OPERATOR(pg_catalog.=) p_identifier_kind
          AND directory.lookup_key_id OPERATOR(pg_catalog.=) p_key_id
          AND directory.identifier_hmac OPERATOR(pg_catalog.=) p_identifier_hmac
          AND directory.state OPERATOR(pg_catalog.=) 'active'
          AND directory.tombstoned_at IS NULL
        FOR SHARE;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'IAM login target is not exact and active';
    END;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_login_target_v1(
    uuid,text,text,bytea
) FROM PUBLIC;
