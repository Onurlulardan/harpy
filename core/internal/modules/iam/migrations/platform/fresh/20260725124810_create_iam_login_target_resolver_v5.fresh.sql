-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_login_target_v5(
    p_tenant_id uuid, p_identifier_kind text,
    p_key_id text, p_identifier_hmac bytea
)
RETURNS TABLE (
    tenant_id uuid, source_user_id uuid, source_identifier_id uuid,
    source_identifier_kind text, source_revision bigint, source_sha256 bytea,
    group_generation bigint, iam_generation bigint, layout_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR
       p_identifier_kind NOT IN ('email_v1','portal_v1') OR
       NOT platform.iam_hmac_key_id_valid_v1(p_key_id) OR
       pg_catalog.octet_length(p_identifier_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM login target V5 input is invalid';
    END IF;
    RETURN QUERY SELECT target.tenant_id, target.source_user_id,
        target.source_identifier_id, target.source_identifier_kind,
        target.source_revision, target.source_sha256,
        target.group_generation, target.iam_generation,
        target.layout_generation
    FROM platform.iam_resolve_login_target_v2(
        p_tenant_id, p_identifier_kind, p_key_id, p_identifier_hmac
    ) AS target;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_login_target_v5(
    uuid,text,text,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v5(
    uuid,text,text,bytea
) TO harpy_runtime;
