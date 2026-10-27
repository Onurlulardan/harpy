-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_seed_v2(
    p_component_id uuid,p_package_id uuid,p_package_version text,
    p_release_id uuid,p_content_sha256 text
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_bytes bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_package_version) NOT BETWEEN 5 AND 128 OR
       p_content_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection seed input is invalid';
    END IF;
    v_bytes:=pg_catalog.substring(pg_catalog.sha256(
        pg_catalog.convert_to('forgecrud:analytics:catalog:projection:v2','UTF8')||
        pg_catalog.uuid_send(p_component_id)||
        pg_catalog.uuid_send(p_package_id)||
        pg_catalog.int4send(pg_catalog.octet_length(p_package_version))||
        pg_catalog.convert_to(p_package_version,'UTF8')||
        pg_catalog.uuid_send(p_release_id)||
        pg_catalog.decode(p_content_sha256,'hex')
    ),1,16);
    v_bytes:=pg_catalog.set_byte(
        v_bytes,6,(pg_catalog.get_byte(v_bytes,6)&15)|112
    );
    v_bytes:=pg_catalog.set_byte(
        v_bytes,8,(pg_catalog.get_byte(v_bytes,8)&63)|128
    );
    RETURN pg_catalog.encode(v_bytes,'hex')::uuid;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_seed_v2(
    uuid,uuid,text,uuid,text
) FROM PUBLIC;
