-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_create_v2(
    p_scope_schema name,
    p_revision_id uuid,
    p_resource_kind text,
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
    v_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_resource_kind NOT IN ('user', 'organization', 'role')
       OR p_revision_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_now IS NULL
       OR NOT system_resource_layout_definition_valid_v2(p_definition_json)
       OR pg_catalog.octet_length(p_definition_sha256) <> 32
       OR p_definition_sha256 IS DISTINCT FROM pg_catalog.sha256(
            pg_catalog.convert_to(p_definition_json, 'UTF8')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_create_v2_contract_guard',
            MESSAGE = 'system resource layout v2 create contract is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'harpy:system_resource_layout:' ||
        pg_catalog.current_schema() || ':' || p_resource_kind, 0
    ));
    SELECT COALESCE(pg_catalog.max(revision.version), 0) + 1
    INTO v_version
    FROM system_resource_layout_revisions AS revision
    WHERE revision.resource_kind = p_resource_kind;
    INSERT INTO system_resource_layout_revisions (
        id, resource_kind, version, state, definition_json,
        definition_sha256, created_at, updated_at, published_at, row_revision
    ) VALUES (
        p_revision_id, p_resource_kind, v_version, 'draft', p_definition_json,
        p_definition_sha256, p_now, p_now, NULL, 1
    );
    RETURN system_resource_layout_revision_json_v1(p_revision_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_create_v2(
    name,uuid,text,text,bytea,timestamptz
) FROM PUBLIC;
