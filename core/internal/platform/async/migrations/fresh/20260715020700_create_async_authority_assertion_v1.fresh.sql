-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_current_authority_v1(
    p_scope_kind text,
    p_scope_id uuid,
    p_group_id text,
    p_group_generation bigint,
    p_domain text,
    p_cutover_generation bigint
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group_state text;
    v_group_writer text;
    v_group_generation bigint;
    v_domain_state text;
    v_domain_writer text;
    v_domain_generation bigint;
BEGIN
    IF p_scope_kind IS NULL OR p_scope_id IS NULL OR p_group_id IS NULL OR
       p_group_generation IS NULL OR p_domain IS NULL OR
       p_cutover_generation IS NULL OR
       NOT (p_scope_kind OPERATOR(pg_catalog.=) ANY (
           ARRAY['global','tenant']::text[]
       )) OR
       p_group_id OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_]{0,62}$' OR
       p_group_generation OPERATOR(pg_catalog.<) 1 OR
       p_domain OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_]{0,62}$' OR
       p_cutover_generation OPERATOR(pg_catalog.<) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async authority input is invalid';
    END IF;

    SELECT group_row.state, group_row.writer, group_row.generation
    INTO v_group_state, v_group_writer, v_group_generation
    FROM platform.cutover_groups AS group_row
    WHERE group_row.scope_kind OPERATOR(pg_catalog.=) p_scope_kind
      AND group_row.scope_id OPERATOR(pg_catalog.=) p_scope_id
      AND group_row.group_id OPERATOR(pg_catalog.=) p_group_id
    FOR SHARE;

    IF NOT FOUND OR v_group_state IS DISTINCT FROM 'v2' OR
       v_group_writer IS DISTINCT FROM 'v2' OR
       v_group_generation IS DISTINCT FROM p_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'async group authority is stale';
    END IF;

    SELECT domain_row.state, domain_row.writer, domain_row.generation
    INTO v_domain_state, v_domain_writer, v_domain_generation
    FROM platform.domain_cutovers AS domain_row
    WHERE domain_row.scope_kind OPERATOR(pg_catalog.=) p_scope_kind
      AND domain_row.scope_id OPERATOR(pg_catalog.=) p_scope_id
      AND domain_row.domain OPERATOR(pg_catalog.=) p_domain
    FOR SHARE;

    IF NOT FOUND OR v_domain_state IS DISTINCT FROM 'v2' OR
       v_domain_writer IS DISTINCT FROM 'v2' OR
       v_domain_generation IS DISTINCT FROM p_cutover_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5602',
            MESSAGE = 'async domain authority is stale';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM platform.cutover_group_members AS member
        WHERE member.scope_kind OPERATOR(pg_catalog.=) p_scope_kind
          AND member.scope_id OPERATOR(pg_catalog.=) p_scope_id
          AND member.group_id OPERATOR(pg_catalog.=) p_group_id
          AND member.domain OPERATOR(pg_catalog.=) p_domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'async authority membership is missing';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_assert_current_authority_v1(
    text,uuid,text,bigint,text,bigint
) FROM PUBLIC;
