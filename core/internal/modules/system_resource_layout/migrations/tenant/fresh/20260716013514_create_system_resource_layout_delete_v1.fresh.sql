-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_delete_v1(
    p_scope_schema name,
    p_revision_id uuid,
    p_expected_row_revision bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_state text;
    v_row_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_expected_row_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_delete_contract_guard',
            MESSAGE = 'system resource layout delete contract is invalid';
    END IF;
    SELECT revision.state, revision.row_revision
    INTO v_state, v_row_revision
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    IF v_state IS DISTINCT FROM 'draft'
       OR v_row_revision IS DISTINCT FROM p_expected_row_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource layout revision conflicts';
    END IF;
    DELETE FROM system_resource_layout_revisions
    WHERE id = p_revision_id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_delete_v1(
    name,uuid,bigint
) FROM PUBLIC;
