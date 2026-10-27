-- +goose Up
DROP FUNCTION platform.contract_retirement_provision_v1(text,text,text,bigint);

-- +goose StatementBegin
CREATE FUNCTION platform.contract_retirement_provision_v1(
    p_contract_key text, p_owner_module text, p_domain text,
    p_remaining_scope_count bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_retirement platform.contract_retirements%ROWTYPE;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_migrator' OR
       p_contract_key IS NULL OR p_owner_module IS NULL OR p_domain IS NULL OR
       p_contract_key OPERATOR(pg_catalog.!~) '^rq[.]alias[.][0-9a-f]{64}$' OR
       p_owner_module OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_]{0,62}$' OR
       NOT (p_domain OPERATOR(pg_catalog.=) ANY (
           ARRAY['analytics','billing','catalog','documents','forms','iam',
                 'notifications','system_resource_layout','work','workflow']::text[]
       )) OR p_remaining_scope_count IS NULL OR
       p_remaining_scope_count OPERATOR(pg_catalog.<) 0 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'contract retirement provision input is invalid';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    INSERT INTO platform.contract_retirements AS retirement (
        contract_key, owner_module, domain, state, remaining_scope_count,
        hit_count, last_hit_at, zero_hit_since, eligible_at, retired_at,
        revision, updated_at
    ) VALUES (
        p_contract_key, p_owner_module, p_domain, 'active',
        p_remaining_scope_count, 0, NULL,
        CASE WHEN p_remaining_scope_count OPERATOR(pg_catalog.=) 0
            THEN v_now ELSE NULL END,
        NULL, NULL, 1, v_now
    ) ON CONFLICT (contract_key) DO NOTHING
    RETURNING retirement.* INTO v_retirement;
    IF NOT FOUND THEN
        SELECT retirement.* INTO v_retirement
        FROM platform.contract_retirements AS retirement
        WHERE retirement.contract_key OPERATOR(pg_catalog.=) p_contract_key
        FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = 'P5501',
                MESSAGE = 'contract retirement key is missing';
        END IF;
    END IF;
    IF v_retirement.owner_module IS DISTINCT FROM p_owner_module OR
       v_retirement.domain IS DISTINCT FROM p_domain THEN
        RAISE EXCEPTION USING ERRCODE = 'P5502',
            MESSAGE = 'contract retirement identity differs';
    END IF;
    IF v_retirement.state OPERATOR(pg_catalog.=) 'retired' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5503',
            MESSAGE = 'retired contract key cannot be provisioned';
    END IF;
    IF v_retirement.remaining_scope_count IS DISTINCT FROM p_remaining_scope_count THEN
        RAISE EXCEPTION USING ERRCODE = 'P5505',
            MESSAGE = 'contract retirement reconciliation is required';
    END IF;
    RETURN v_retirement.revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.contract_retirement_provision_v1(
    text,text,text,bigint
) FROM PUBLIC;
