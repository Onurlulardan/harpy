-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_login_target_v2(
    p_tenant_id uuid,
    p_identifier_kind text,
    p_key_id text,
    p_identifier_hmac bytea
)
RETURNS TABLE (
    tenant_id uuid,
    source_user_id uuid,
    source_identifier_id uuid,
    source_identifier_kind text,
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
DECLARE
    v_tenant_id uuid;
    v_source_revision bigint;
    v_source_sha256 bytea;
BEGIN
    SELECT target.tenant_id, target.source_revision, target.source_sha256,
           target.group_generation, target.iam_generation,
           target.layout_generation
    INTO STRICT v_tenant_id, v_source_revision, v_source_sha256,
        group_generation, iam_generation, layout_generation
    FROM platform.iam_resolve_login_target_v1(
        p_tenant_id, p_identifier_kind, p_key_id, p_identifier_hmac
    ) AS target;

    SELECT directory.tenant_id, directory.user_id,
           directory.source_identifier_id, directory.identifier_kind,
           directory.source_revision, directory.source_sha256
    INTO STRICT tenant_id, source_user_id, source_identifier_id,
        source_identifier_kind, source_revision, source_sha256
    FROM platform.iam_login_directory AS directory
    WHERE directory.tenant_id OPERATOR(pg_catalog.=) v_tenant_id
      AND directory.identifier_kind OPERATOR(pg_catalog.=) p_identifier_kind
      AND directory.lookup_key_id OPERATOR(pg_catalog.=) p_key_id
      AND directory.identifier_hmac OPERATOR(pg_catalog.=) p_identifier_hmac
      AND directory.source_revision OPERATOR(pg_catalog.=) v_source_revision
      AND directory.source_sha256 OPERATOR(pg_catalog.=) v_source_sha256
      AND directory.state OPERATOR(pg_catalog.=) 'active'
      AND directory.tombstoned_at IS NULL
    FOR SHARE;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_login_target_v2(
    uuid,text,text,bytea
) FROM PUBLIC;
