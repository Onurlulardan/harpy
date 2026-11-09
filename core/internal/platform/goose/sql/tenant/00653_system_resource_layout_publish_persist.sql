-- +goose Up
DROP FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz);

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v5(
    in_schema_name name, in_id uuid, in_expected_row_revision bigint, in_source_sha256 bytea, in_compiler_version text,
    in_catalog_contract text, in_catalog_json text, in_catalog_sha256 bytea, in_presentation_contract text,
    in_presentation_json text, in_presentation_sha256 bytea, in_now timestamptz
) RETURNS bytea LANGUAGE plpgsql VOLATILE SECURITY DEFINER
AS $fn$
DECLARE
    tenant_schema name; definition text; stored_sha text; state text; row_revision bigint; created_at timestamptz;
    publication text; stored_catalog text; source_sha text; catalog_sha text; presentation_sha text; revision_document text; resource_kind text;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    source_sha := pg_catalog.encode(in_source_sha256, 'hex');
    catalog_sha := pg_catalog.encode(in_catalog_sha256, 'hex');
    presentation_sha := pg_catalog.encode(in_presentation_sha256, 'hex');
    IF tenant_schema IS DISTINCT FROM in_schema_name OR in_expected_row_revision < 1 OR in_now IS NULL
        OR in_catalog_json IS NULL OR in_presentation_json IS NULL THEN
        RAISE EXCEPTION 'system resource publication is invalid' USING ERRCODE = '22023';
    END IF;
    EXECUTE pg_catalog.format('SELECT resource_kind, definition_json, definition_sha256, state, row_revision, created_at, publication_document, catalog_sha256 FROM %I.system_resource_layout_revisions WHERE id = $1', tenant_schema)
        INTO resource_kind, definition, stored_sha, state, row_revision, created_at, publication, stored_catalog USING in_id;
    IF definition IS NULL THEN RAISE EXCEPTION 'system resource revision was not found' USING ERRCODE = 'P0002'; END IF;
    IF stored_sha IS DISTINCT FROM source_sha THEN RAISE EXCEPTION 'system resource source digest does not match' USING ERRCODE = '23514'; END IF;
    IF state = 'published' AND row_revision = in_expected_row_revision + 1 AND stored_catalog = catalog_sha AND publication IS NOT NULL THEN
        RETURN pg_catalog.convert_to(publication, 'UTF8');
    END IF;
    IF state <> 'draft' OR row_revision <> in_expected_row_revision THEN
        RAISE EXCEPTION 'system resource revision conflicts' USING ERRCODE = '40001';
    END IF;
    EXECUTE pg_catalog.format('SELECT %I.system_resource_layout_revision_document_v1($1,$2,1,''published'',$3,$4,$5,$6,$6,$7)', tenant_schema)
        INTO revision_document USING in_id, resource_kind, definition, stored_sha, created_at, in_now, in_expected_row_revision + 1;
    publication := '{"revision":' || revision_document || ',"generation":1'
        || ',"source_definition_sha256":' || pg_catalog.to_json(source_sha)
        || ',"compiler_version":' || pg_catalog.to_json(in_compiler_version)
        || ',"catalog_contract":' || pg_catalog.to_json(in_catalog_contract)
        || ',"catalog_json":' || pg_catalog.to_json(in_catalog_json)
        || ',"catalog_sha256":' || pg_catalog.to_json(catalog_sha)
        || ',"presentation_contract":' || pg_catalog.to_json(in_presentation_contract)
        || ',"presentation_json":' || pg_catalog.to_json(in_presentation_json)
        || ',"presentation_sha256":' || pg_catalog.to_json(presentation_sha) || '}';
    EXECUTE pg_catalog.format('UPDATE %I.system_resource_layout_revisions SET document = $2, state = ''published'', row_revision = $3, updated_at = $4, published_at = $4, publication_document = $5, catalog_sha256 = $6 WHERE id = $1 AND state = ''draft'' AND row_revision = $7', tenant_schema)
        USING in_id, revision_document, in_expected_row_revision + 1, in_now, publication, catalog_sha, in_expected_row_revision;
    RETURN pg_catalog.convert_to(publication, 'UTF8');
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz);
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v5(arg_1 name, arg_2 uuid, arg_3 bigint, arg_4 bytea, arg_5 text, arg_6 text, arg_7 text, arg_8 bytea, arg_9 text, arg_10 text, arg_11 bytea, arg_12 timestamptz)
RETURNS bytea LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT NULL::bytea; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;
