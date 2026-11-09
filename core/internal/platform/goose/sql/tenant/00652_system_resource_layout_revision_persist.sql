-- +goose Up
CREATE TABLE IF NOT EXISTS system_resource_layout_revisions (
    id uuid PRIMARY KEY,
    resource_kind text NOT NULL,
    document text NOT NULL,
    definition_json text NOT NULL,
    definition_sha256 text NOT NULL,
    state text NOT NULL,
    row_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    published_at timestamptz,
    publication_document text,
    catalog_sha256 text
);

REVOKE ALL ON TABLE system_resource_layout_revisions FROM PUBLIC, harpy_api, harpy_worker;

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_revision_document_v1(
    in_id uuid, in_kind text, in_version bigint, in_state text, in_definition text, in_sha text,
    in_created timestamptz, in_updated timestamptz, in_published timestamptz, in_row_revision bigint
) RETURNS text LANGUAGE sql IMMUTABLE SET search_path TO pg_catalog
AS $fn$
    SELECT '{' || '"id":' || pg_catalog.to_json(in_id::text) || ',"resource_kind":' || pg_catalog.to_json(in_kind)
        || ',"version":' || in_version::text || ',"state":' || pg_catalog.to_json(in_state) || ',"definition":' || in_definition
        || ',"definition_sha256":' || pg_catalog.to_json(in_sha)
        || ',"created_at":' || pg_catalog.to_json(pg_catalog.to_jsonb(in_created) #>> '{}')
        || ',"updated_at":' || pg_catalog.to_json(pg_catalog.to_jsonb(in_updated) #>> '{}')
        || ',"published_at":' || CASE WHEN in_published IS NULL THEN 'null' ELSE pg_catalog.to_json(pg_catalog.to_jsonb(in_published) #>> '{}') END
        || ',"row_revision":' || in_row_revision::text || '}';
$fn$;
-- +goose StatementEnd

DROP FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz);
DROP FUNCTION system_resource_layout_get_v1(name, uuid);

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_create_v3(in_schema_name name, in_id uuid, in_resource text, in_definition text, in_definition_sha256 bytea, in_now timestamptz)
RETURNS bytea LANGUAGE plpgsql VOLATILE SECURITY DEFINER
AS $fn$
DECLARE tenant_schema name; document text; sha text;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    sha := pg_catalog.encode(in_definition_sha256, 'hex');
    IF tenant_schema IS DISTINCT FROM in_schema_name OR in_resource NOT IN ('user', 'organization', 'role')
        OR in_definition IS NULL OR pg_catalog.left(in_definition, 1) <> '{' OR pg_catalog.octet_length(in_definition_sha256) <> 32 OR in_now IS NULL THEN
        RAISE EXCEPTION 'system resource revision is invalid' USING ERRCODE = '22023';
    END IF;
    PERFORM in_definition::jsonb;
    EXECUTE pg_catalog.format('SELECT %I.system_resource_layout_revision_document_v1($1,$2,1,''draft'',$3,$4,$5,$5,NULL,1)', tenant_schema)
        INTO document USING in_id, in_resource, in_definition, sha, in_now;
    EXECUTE pg_catalog.format('INSERT INTO %I.system_resource_layout_revisions (id, resource_kind, document, definition_json, definition_sha256, state, row_revision, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,''draft'',1,$6,$6) ON CONFLICT (id) DO NOTHING', tenant_schema)
        USING in_id, in_resource, document, in_definition, sha, in_now;
    EXECUTE pg_catalog.format('SELECT document FROM %I.system_resource_layout_revisions WHERE id = $1 AND resource_kind = $2 AND definition_sha256 = $3 AND state = ''draft'' AND row_revision = 1', tenant_schema)
        INTO document USING in_id, in_resource, sha;
    IF document IS NULL THEN RAISE EXCEPTION 'system resource revision conflicts' USING ERRCODE = '40001'; END IF;
    RETURN pg_catalog.convert_to(document, 'UTF8');
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_get_v1(in_schema_name name, in_id uuid)
RETURNS bytea LANGUAGE plpgsql STABLE SECURITY DEFINER
AS $fn$
DECLARE tenant_schema name; document text;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    IF tenant_schema IS DISTINCT FROM in_schema_name THEN RAISE EXCEPTION 'system resource schema is not bound' USING ERRCODE = '42501'; END IF;
    EXECUTE pg_catalog.format('SELECT document FROM %I.system_resource_layout_revisions WHERE id = $1', tenant_schema) INTO document USING in_id;
    IF document IS NULL THEN RAISE EXCEPTION 'system resource revision was not found' USING ERRCODE = 'P0002'; END IF;
    RETURN pg_catalog.convert_to(document, 'UTF8');
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_revision_document_v1(uuid, text, bigint, text, text, text, timestamptz, timestamptz, timestamptz, bigint) FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_get_v1(name, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION system_resource_layout_get_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_get_v1(name, uuid);
DROP FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz);
DROP FUNCTION system_resource_layout_revision_document_v1(uuid, text, bigint, text, text, text, timestamptz, timestamptz, timestamptz, bigint);
DROP TABLE IF EXISTS system_resource_layout_revisions;
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_create_v3(arg_1 name, arg_2 uuid, arg_3 text, arg_4 text, arg_5 bytea, arg_6 timestamptz)
RETURNS bytea LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT NULL::bytea; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_get_v1(arg_1 name, arg_2 uuid)
RETURNS bytea LANGUAGE sql STABLE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT NULL::bytea; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_get_v1(name, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION system_resource_layout_get_v1(name, uuid) TO harpy_api, harpy_worker;
