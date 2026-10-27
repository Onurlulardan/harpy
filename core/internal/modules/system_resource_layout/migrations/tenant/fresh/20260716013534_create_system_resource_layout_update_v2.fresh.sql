-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_update_v2(
    p_scope_schema name,
    p_revision_id uuid,
    p_expected_row_revision bigint,
    p_definition_json text,
    p_definition_sha256 bytea,
    p_now timestamptz
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_contract text;
    v_row_revision bigint;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_expected_row_revision < 1 OR p_now IS NULL
       OR NOT system_resource_layout_definition_valid_v2(p_definition_json)
       OR pg_catalog.octet_length(p_definition_sha256) <> 32
       OR p_definition_sha256 IS DISTINCT FROM pg_catalog.sha256(
            pg_catalog.convert_to(p_definition_json, 'UTF8')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_update_v2_contract_guard',
            MESSAGE = 'system resource layout v2 update contract is invalid';
    END IF;
    SELECT revision.state, revision.row_revision,
           revision.definition_json::jsonb ->> 'contract'
    INTO v_state, v_row_revision, v_contract
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    IF v_state IS DISTINCT FROM 'draft'
       OR v_row_revision IS DISTINCT FROM p_expected_row_revision
       OR v_contract IS DISTINCT FROM 'forgecrud/system-resource-layout/v2' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource layout revision conflicts';
    END IF;
    UPDATE system_resource_layout_revisions
    SET definition_json = p_definition_json,
        definition_sha256 = p_definition_sha256,
        updated_at = p_now,
        row_revision = row_revision + 1
    WHERE id = p_revision_id;
    RETURN system_resource_layout_revision_json_v1(p_revision_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_update_v2(
    name,uuid,bigint,text,bytea,timestamptz
) FROM PUBLIC;
