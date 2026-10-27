-- +goose Up
-- Migration stream version: 20260730120100.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_canonical_json_v1(p_value jsonb)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_result text;
BEGIN
    v_kind := pg_catalog.jsonb_typeof(p_value);
    IF v_kind = 'object' THEN
        SELECT '{' || COALESCE(pg_catalog.string_agg(
            pg_catalog.to_json(entry.key)::text || ':' ||
            system_resource_layout_canonical_json_v1(entry.value),
            ',' ORDER BY entry.key COLLATE "C"
        ), '') || '}'
        INTO v_result
        FROM pg_catalog.jsonb_each(p_value) AS entry(key, value);
        RETURN v_result;
    ELSIF v_kind = 'array' THEN
        SELECT '[' || COALESCE(pg_catalog.string_agg(
            system_resource_layout_canonical_json_v1(entry.value),
            ',' ORDER BY entry.ordinality
        ), '') || ']'
        INTO v_result
        FROM pg_catalog.jsonb_array_elements(p_value)
            WITH ORDINALITY AS entry(value, ordinality);
        RETURN v_result;
    ELSIF v_kind IN ('string', 'number', 'boolean', 'null') THEN
        RETURN p_value::text;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '22023',
        MESSAGE = 'system resource layout canonical JSON input is invalid';
END;
$function$;

CREATE FUNCTION system_resource_layout_deterministic_uuid_v7_v1(
    p_seed text, p_path text, p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_milliseconds bigint;
    v_digest bytea;
    v_value bytea;
    v_hex text;
BEGIN
    IF p_seed IS NULL OR p_seed = '' OR p_path IS NULL OR p_path = ''
       OR p_occurred_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'system resource layout deterministic identity is invalid';
    END IF;
    v_milliseconds := pg_catalog.floor(
        pg_catalog.date_part('epoch', p_occurred_at) * 1000
    )::bigint;
    IF v_milliseconds < 0 OR v_milliseconds >= 281474976710656 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'system resource layout deterministic time is invalid';
    END IF;
    v_digest := pg_catalog.sha256(
        pg_catalog.convert_to(p_seed, 'UTF8') ||
        pg_catalog.decode('00', 'hex') ||
        pg_catalog.convert_to(p_path, 'UTF8')
    );
    v_value := substring(
        pg_catalog.int8send(v_milliseconds) FROM 3 FOR 6
    ) || substring(v_digest FROM 1 FOR 10);
    v_value := pg_catalog.set_byte(
        v_value, 6, (pg_catalog.get_byte(v_value, 6) & 15) | 112
    );
    v_value := pg_catalog.set_byte(
        v_value, 8, (pg_catalog.get_byte(v_value, 8) & 63) | 128
    );
    v_hex := pg_catalog.encode(v_value, 'hex');
    RETURN (
        substring(v_hex FROM 1 FOR 8) || '-' ||
        substring(v_hex FROM 9 FOR 4) || '-' ||
        substring(v_hex FROM 13 FOR 4) || '-' ||
        substring(v_hex FROM 17 FOR 4) || '-' ||
        substring(v_hex FROM 21 FOR 12)
    )::uuid;
END;
$function$;

-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_canonical_json_v1(jsonb) FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_deterministic_uuid_v7_v1(
    text,text,timestamptz) FROM PUBLIC;
