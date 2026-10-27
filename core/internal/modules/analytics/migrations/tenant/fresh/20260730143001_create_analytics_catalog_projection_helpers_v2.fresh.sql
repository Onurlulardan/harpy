-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_uuid_v7_v1(
    p_component_id uuid, p_label text
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_bytes bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_label IS NULL OR
       pg_catalog.octet_length(p_label) NOT BETWEEN 1 AND 100 OR
       p_label OPERATOR(pg_catalog.!~) '^[a-z0-9:-]+$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog deterministic identity input is invalid';
    END IF;
    v_bytes := pg_catalog.substring(pg_catalog.sha256(
        pg_catalog.convert_to('harpy:analytics:catalog:v2','UTF8') ||
        pg_catalog.uuid_send(p_component_id) ||
        pg_catalog.decode('00','hex') ||
        pg_catalog.convert_to(p_label,'UTF8')
    ), 1, 16);
    v_bytes := pg_catalog.set_byte(
        v_bytes, 6, (pg_catalog.get_byte(v_bytes, 6) & 15) | 112
    );
    v_bytes := pg_catalog.set_byte(
        v_bytes, 8, (pg_catalog.get_byte(v_bytes, 8) & 63) | 128
    );
    RETURN pg_catalog.encode(v_bytes,'hex')::uuid;
END;
$function$;

CREATE FUNCTION analytics_catalog_dependency_manifest_v1(
    p_resource_id uuid, p_resource_version_id uuid,
    p_catalog_version bigint, p_query_contract_version bigint,
    p_policy_version bigint, p_data_revision_capability text,
    p_dependency_sha256 bytea
)
RETURNS bytea
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_resource_id IS NULL OR p_resource_version_id IS NULL OR
       p_catalog_version < 1 OR p_query_contract_version < 1 OR
       p_policy_version < 1 OR
       p_data_revision_capability IS NULL OR
       pg_catalog.octet_length(p_data_revision_capability) NOT BETWEEN 0 AND 128 OR
       pg_catalog.octet_length(p_dependency_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog dependency manifest input is invalid';
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.int4send(3) || pg_catalog.convert_to('iam','UTF8') ||
        pg_catalog.int4send(4) || pg_catalog.convert_to('form','UTF8') ||
        pg_catalog.int4send(16) || pg_catalog.uuid_send(p_resource_id) ||
        pg_catalog.int4send(16) || pg_catalog.uuid_send(p_resource_version_id) ||
        pg_catalog.int4send(pg_catalog.octet_length(p_data_revision_capability)) ||
        pg_catalog.convert_to(p_data_revision_capability,'UTF8') ||
        pg_catalog.int4send(32) || p_dependency_sha256 ||
        pg_catalog.int8send(p_catalog_version) ||
        pg_catalog.int8send(p_query_contract_version) ||
        pg_catalog.int8send(p_policy_version)
    );
END;
$function$;

CREATE FUNCTION analytics_catalog_dependencies_manifest_v1(
    p_dependencies jsonb
)
RETURNS bytea
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_dependency jsonb;
    v_material bytea;
    v_owner text;
    v_kind text;
    v_resource_id uuid;
    v_version_id uuid;
    v_catalog bigint;
    v_query bigint;
    v_policy bigint;
    v_revision text;
    v_digest bytea;
    v_identity text;
    v_previous text;
BEGIN
    v_material := ''::bytea;
    IF pg_catalog.jsonb_typeof(p_dependencies) IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_dependencies) NOT BETWEEN 1 AND 4 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog dependencies are invalid';
    END IF;
    FOR v_dependency IN
        SELECT item.value FROM pg_catalog.jsonb_array_elements(p_dependencies)
            AS item(value)
        ORDER BY item.value->>'owner_domain',item.value->>'resource_kind',
            item.value->>'resource_id',item.value->>'resource_version_id'
    LOOP
        IF v_dependency->>'owner_domain' <> 'iam' OR
           v_dependency->>'resource_kind' <> 'form' OR
           NOT pg_catalog.pg_input_is_valid(
               v_dependency->>'resource_id','uuid') OR
           NOT pg_catalog.pg_input_is_valid(
               v_dependency->>'resource_version_id','uuid') OR
           v_dependency->>'dependency_sha256'
               OPERATOR(pg_catalog.!~) '^[0-9a-f]{64}$' THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog dependency is invalid';
        END IF;
        v_owner := v_dependency->>'owner_domain';
        v_kind := v_dependency->>'resource_kind';
        v_resource_id := (v_dependency->>'resource_id')::uuid;
        v_version_id := (v_dependency->>'resource_version_id')::uuid;
        v_catalog := (v_dependency->>'catalog_version')::bigint;
        v_query := (v_dependency->>'query_contract_version')::bigint;
        v_policy := (v_dependency->>'policy_version')::bigint;
        v_revision := v_dependency->>'data_revision_capability';
        v_digest := pg_catalog.decode(
            v_dependency->>'dependency_sha256','hex'
        );
        v_identity := v_owner||'|'||v_kind||'|'||
            v_resource_id::text||'|'||v_version_id::text;
        IF v_identity IS NOT DISTINCT FROM v_previous OR
           pg_catalog.uuid_extract_version(v_resource_id) IS DISTINCT FROM 7 OR
           pg_catalog.uuid_extract_version(v_version_id) IS DISTINCT FROM 7 OR
           v_catalog < 1 OR v_query < 1 OR v_policy < 1 OR
           v_revision IS NULL OR
           pg_catalog.octet_length(v_revision) NOT BETWEEN 0 AND 128 OR
           pg_catalog.octet_length(v_digest) IS DISTINCT FROM 32 THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog dependency differs';
        END IF;
        v_previous := v_identity;
        v_material := v_material ||
            pg_catalog.int4send(pg_catalog.octet_length(v_owner)) ||
            pg_catalog.convert_to(v_owner,'UTF8') ||
            pg_catalog.int4send(pg_catalog.octet_length(v_kind)) ||
            pg_catalog.convert_to(v_kind,'UTF8') ||
            pg_catalog.int4send(16) || pg_catalog.uuid_send(v_resource_id) ||
            pg_catalog.int4send(16) || pg_catalog.uuid_send(v_version_id) ||
            pg_catalog.int4send(pg_catalog.octet_length(v_revision)) ||
            pg_catalog.convert_to(v_revision,'UTF8') ||
            pg_catalog.int4send(32) || v_digest ||
            pg_catalog.int8send(v_catalog) ||
            pg_catalog.int8send(v_query) ||
            pg_catalog.int8send(v_policy);
    END LOOP;
    RETURN pg_catalog.sha256(v_material);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_uuid_v7_v1(uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_dependency_manifest_v1(
    uuid,uuid,bigint,bigint,bigint,text,bytea
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_dependencies_manifest_v1(jsonb)
FROM PUBLIC;
