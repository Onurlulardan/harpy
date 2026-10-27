-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_export_maintenance_effect_id_v1(
    p_attempt_id uuid,p_kind text
)
RETURNS uuid
LANGUAGE plpgsql IMMUTABLE SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_namespace uuid;
    v_digest bytea;
    v_uuid bytea;
    v_hex text;
BEGIN
    v_namespace:=CASE p_kind
        WHEN 'analytics.dashboard.export.delete' THEN
            'bfce9d42-fd8b-453a-a54e-d6f35a886868'::uuid
        WHEN 'analytics.dashboard.export.expire' THEN
            '1bce05e3-fc4d-40a9-82e9-f125b12fa10a'::uuid END;
    IF pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       v_namespace IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance effect identity input is invalid';
    END IF;
    v_digest:=platform.platform_async_analytics_export_sha1_v1(
        pg_catalog.uuid_send(v_namespace)||pg_catalog.uuid_send(p_attempt_id));
    v_uuid:=pg_catalog.substr(v_digest,1,16);
    v_uuid:=pg_catalog.set_byte(v_uuid,6,
        (pg_catalog.get_byte(v_uuid,6)&15)|80);
    v_uuid:=pg_catalog.set_byte(v_uuid,8,
        (pg_catalog.get_byte(v_uuid,8)&63)|128);
    v_hex:=pg_catalog.encode(v_uuid,'hex');
    RETURN (pg_catalog.substr(v_hex,1,8)||'-'||pg_catalog.substr(v_hex,9,4)||'-'||
        pg_catalog.substr(v_hex,13,4)||'-'||pg_catalog.substr(v_hex,17,4)||'-'||
        pg_catalog.substr(v_hex,21,12))::uuid;
END;
$function$;

CREATE FUNCTION platform.platform_async_analytics_export_maintenance_receipt_id_v2(
    p_effect_id uuid,p_code text
)
RETURNS uuid
LANGUAGE plpgsql IMMUTABLE SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_code text;
    v_digest bytea;
    v_uuid bytea;
    v_hex text;
BEGIN
    v_code:=CASE p_code
        WHEN 'ANALYTICS_EXPORT_DELETED' THEN p_code
        WHEN 'ANALYTICS_EXPORT_EXPIRY_COMPLETE' THEN p_code
        WHEN 'ANALYTICS_EXPORT_REJECTED' THEN p_code
        WHEN 'ANALYTICS_EXPORT_CONFLICT' THEN p_code
        WHEN 'ANALYTICS_EXPORT_UNAVAILABLE' THEN p_code
        WHEN 'ANALYTICS_EXPORT_RETRY' THEN p_code
        WHEN 'ANALYTICS_EXPORT_TERMINAL' THEN p_code END;
    IF pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 5 OR
       v_code IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance receipt identity input is invalid';
    END IF;
    v_digest:=platform.platform_async_analytics_export_sha1_v1(
        pg_catalog.uuid_send('8e7a6d02-df6d-4327-b179-160296b50b8e'::uuid)||
        pg_catalog.convert_to(p_effect_id::text,'UTF8')||
        pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_code,'UTF8'));
    v_uuid:=pg_catalog.substr(v_digest,1,16);
    v_uuid:=pg_catalog.set_byte(v_uuid,6,
        (pg_catalog.get_byte(v_uuid,6)&15)|80);
    v_uuid:=pg_catalog.set_byte(v_uuid,8,
        (pg_catalog.get_byte(v_uuid,8)&63)|128);
    v_hex:=pg_catalog.encode(v_uuid,'hex');
    RETURN (pg_catalog.substr(v_hex,1,8)||'-'||pg_catalog.substr(v_hex,9,4)||'-'||
        pg_catalog.substr(v_hex,13,4)||'-'||pg_catalog.substr(v_hex,17,4)||'-'||
        pg_catalog.substr(v_hex,21,12))::uuid;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_export_maintenance_effect_id_v1(uuid,text)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_export_maintenance_receipt_id_v2(uuid,text)
FROM PUBLIC;
