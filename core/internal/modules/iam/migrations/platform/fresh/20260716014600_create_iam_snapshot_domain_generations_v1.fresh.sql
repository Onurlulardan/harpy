-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_snapshot_domain_generations_v1(
    p_tenant_id uuid,
    p_first_domain text,
    p_second_domain text
)
RETURNS TABLE (result_domain text, result_generation bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR
       p_first_domain IS NULL OR p_second_domain IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM snapshot authority caller is invalid';
    END IF;
    IF p_first_domain !~ '^[a-z][a-z0-9_]{0,62}$' OR
       p_second_domain !~ '^[a-z][a-z0-9_]{0,62}$' OR
       p_first_domain = p_second_domain THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM snapshot authority domains are invalid';
    END IF;
    RETURN QUERY
    SELECT authority.domain, authority.generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.domain IN (p_first_domain,p_second_domain)
      AND authority.state = 'v2' AND authority.writer = 'v2'
    ORDER BY authority.domain;
    IF (SELECT pg_catalog.count(*) FROM platform.domain_cutovers AS authority
        WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
          AND authority.domain IN (p_first_domain,p_second_domain)
          AND authority.state = 'v2' AND authority.writer = 'v2') <> 2 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM snapshot authority is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_snapshot_domain_generations_v1(
    uuid,text,text
) FROM PUBLIC;
