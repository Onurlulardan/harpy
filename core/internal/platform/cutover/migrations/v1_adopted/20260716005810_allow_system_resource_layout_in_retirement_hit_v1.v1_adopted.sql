-- +goose Up
DROP FUNCTION platform.contract_retirement_hit_v1(jsonb,text,text);

-- +goose StatementBegin
CREATE FUNCTION platform.contract_retirement_hit_v1(
    p_contract_keys jsonb, p_owner_module text, p_domain text
)
RETURNS TABLE (retired_contract_key text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_input_count bigint;
    v_distinct_count bigint;
    v_found_count bigint;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_contract_keys IS NULL OR p_owner_module IS NULL OR p_domain IS NULL OR
       p_owner_module OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_]{0,62}$' OR
       NOT (p_domain OPERATOR(pg_catalog.=) ANY (
           ARRAY['analytics','billing','catalog','documents','forms','iam',
                 'notifications','system_resource_layout','work','workflow']::text[]
       )) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement hit input is invalid';
    END IF;
    IF pg_catalog.jsonb_typeof(p_contract_keys) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement hit input is invalid';
    END IF;
    IF pg_catalog.jsonb_array_length(p_contract_keys) OPERATOR(pg_catalog.<) 1 OR
       pg_catalog.jsonb_array_length(p_contract_keys) OPERATOR(pg_catalog.>) 256 OR
       EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_contract_keys) AS item(value)
           WHERE pg_catalog.jsonb_typeof(item.value) IS DISTINCT FROM 'string'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement hit input is invalid';
    END IF;
    SELECT pg_catalog.count(*), pg_catalog.count(DISTINCT key_value)
    INTO STRICT v_input_count, v_distinct_count
    FROM pg_catalog.jsonb_array_elements_text(p_contract_keys) AS key_row(key_value)
    WHERE key_value OPERATOR(pg_catalog.~) '^rq[.]alias[.][0-9a-f]{64}$';
    IF v_input_count IS DISTINCT FROM
           pg_catalog.jsonb_array_length(p_contract_keys)::pg_catalog.int8 OR
       v_distinct_count IS DISTINCT FROM v_input_count THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement hit keys are invalid';
    END IF;
    PERFORM retirement.contract_key
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
        SELECT key_value
        FROM pg_catalog.jsonb_array_elements_text(p_contract_keys) AS key_row(key_value)
    )
    ORDER BY retirement.contract_key COLLATE "C"
    FOR UPDATE;
    SELECT pg_catalog.count(*) INTO STRICT v_found_count
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
        SELECT key_value
        FROM pg_catalog.jsonb_array_elements_text(p_contract_keys) AS key_row(key_value)
    );
    IF v_found_count IS DISTINCT FROM v_input_count THEN
        RAISE EXCEPTION USING ERRCODE = 'P5501',
            MESSAGE = 'contract retirement key is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.contract_retirements AS retirement
        WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
            SELECT key_value FROM pg_catalog.jsonb_array_elements_text(p_contract_keys)
                AS key_row(key_value)
        ) AND (retirement.owner_module IS DISTINCT FROM p_owner_module OR
               retirement.domain IS DISTINCT FROM p_domain)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5502',
            MESSAGE = 'contract retirement identity differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.contract_retirements AS retirement
        WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
            SELECT key_value FROM pg_catalog.jsonb_array_elements_text(p_contract_keys)
                AS key_row(key_value)
        ) AND retirement.state OPERATOR(pg_catalog.=) 'retired'
    ) THEN
        RETURN QUERY
        SELECT retirement.contract_key
        FROM platform.contract_retirements AS retirement
        WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
            SELECT key_value FROM pg_catalog.jsonb_array_elements_text(p_contract_keys)
                AS key_row(key_value)
        ) AND retirement.state OPERATOR(pg_catalog.=) 'retired'
        ORDER BY retirement.contract_key COLLATE "C";
        RETURN;
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.contract_retirements AS retirement
        WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
            SELECT key_value FROM pg_catalog.jsonb_array_elements_text(p_contract_keys)
                AS key_row(key_value)
        ) AND (
            retirement.hit_count OPERATOR(pg_catalog.=) 9223372036854775807 OR
            retirement.revision OPERATOR(pg_catalog.=) 9223372036854775807
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5504',
            MESSAGE = 'contract retirement counter overflow';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    UPDATE platform.contract_retirements AS retirement
    SET hit_count = retirement.hit_count OPERATOR(pg_catalog.+) 1,
        revision = retirement.revision OPERATOR(pg_catalog.+) 1,
        last_hit_at = v_now, updated_at = v_now, state = 'active',
        zero_hit_since = NULL, eligible_at = NULL, retired_at = NULL
    WHERE retirement.contract_key OPERATOR(pg_catalog.=) ANY (
        SELECT key_value
        FROM pg_catalog.jsonb_array_elements_text(p_contract_keys) AS key_row(key_value)
    );
    RETURN;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.contract_retirement_hit_v1(jsonb,text,text) FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.contract_retirement_hit_v1(jsonb,text,text)
TO harpy_runtime;
