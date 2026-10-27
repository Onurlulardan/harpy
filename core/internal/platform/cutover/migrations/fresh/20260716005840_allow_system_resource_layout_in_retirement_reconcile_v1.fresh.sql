-- +goose Up
DROP FUNCTION platform.contract_retirement_reconcile_v1(text,text,text,bigint,bigint);

-- +goose StatementBegin
CREATE FUNCTION platform.contract_retirement_reconcile_v1(
    p_contract_key text, p_owner_module text, p_domain text,
    p_remaining_scope_count bigint, p_expected_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_retirement platform.contract_retirements%ROWTYPE;
    v_now timestamptz;
    v_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator' OR
       p_contract_key IS NULL OR p_owner_module IS NULL OR p_domain IS NULL OR
       p_contract_key OPERATOR(pg_catalog.!~) '^rq[.]alias[.][0-9a-f]{64}$' OR
       p_owner_module OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_]{0,62}$' OR
       NOT (p_domain OPERATOR(pg_catalog.=) ANY (
           ARRAY['analytics','billing','catalog','documents','iam',
                 'notifications','system_resource_layout','work','workflow']::text[]
       )) OR
       p_remaining_scope_count IS NULL OR
       p_remaining_scope_count OPERATOR(pg_catalog.<) 0 OR
       p_expected_revision IS NULL OR p_expected_revision OPERATOR(pg_catalog.<) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement reconcile input is invalid';
    END IF;
    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key OPERATOR(pg_catalog.=) p_contract_key
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5501',
            MESSAGE = 'contract retirement key is missing';
    END IF;
    IF v_retirement.owner_module IS DISTINCT FROM p_owner_module OR
       v_retirement.domain IS DISTINCT FROM p_domain THEN
        RAISE EXCEPTION USING ERRCODE = 'P5502',
            MESSAGE = 'contract retirement identity differs';
    END IF;
    IF v_retirement.state OPERATOR(pg_catalog.=) 'retired' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5503',
            MESSAGE = 'retired contract key cannot be reconciled';
    END IF;
    IF v_retirement.revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = 'P5505',
            MESSAGE = 'contract retirement revision differs';
    END IF;
    IF v_retirement.revision OPERATOR(pg_catalog.=) 9223372036854775807 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5504',
            MESSAGE = 'contract retirement counter overflow';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    UPDATE platform.contract_retirements AS retirement
    SET remaining_scope_count = p_remaining_scope_count,
        state = CASE
            WHEN p_remaining_scope_count OPERATOR(pg_catalog.>) 0 THEN 'active'
            WHEN retirement.state OPERATOR(pg_catalog.=) 'eligible' THEN 'eligible'
            ELSE 'active'
        END,
        zero_hit_since = CASE
            WHEN p_remaining_scope_count OPERATOR(pg_catalog.>) 0 THEN NULL
            ELSE COALESCE(retirement.zero_hit_since, v_now)
        END,
        eligible_at = CASE
            WHEN p_remaining_scope_count OPERATOR(pg_catalog.=) 0 AND
                 retirement.state OPERATOR(pg_catalog.=) 'eligible'
                THEN retirement.eligible_at
            ELSE NULL
        END,
        revision = retirement.revision OPERATOR(pg_catalog.+) 1,
        updated_at = v_now
    WHERE retirement.contract_key OPERATOR(pg_catalog.=) p_contract_key
    RETURNING retirement.revision INTO STRICT v_revision;
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.contract_retirement_reconcile_v1(
    text,text,text,bigint,bigint
) FROM PUBLIC;
