-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_plan_register_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_run_id uuid,
    p_tenant_id uuid,
    p_plan_bytes bytea,
    p_plan_sha256 text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_plan jsonb;
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_issue_count integer;
    v_eligible boolean;
    v_existing bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 adoption plan principal or scope denied';
    END IF;
    IF p_operation_id IS NULL OR p_run_id IS NULL OR p_tenant_id IS NULL OR
       p_plan_sha256 !~ '^[0-9a-f]{64}$' OR
       pg_catalog.octet_length(p_plan_bytes) NOT BETWEEN 2 AND 8388608 OR
       pg_catalog.encode(pg_catalog.sha256(p_plan_bytes), 'hex')
           IS DISTINCT FROM p_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption plan evidence is invalid';
    END IF;
    v_plan := pg_catalog.convert_from(p_plan_bytes, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_plan) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_plan) AS key)
       IS DISTINCT FROM ARRAY[
           'categories', 'contract_version', 'eligible', 'issues',
           'mapping_sha256', 'source_sha256', 'tenant_id'
       ]::text[] OR
       v_plan->>'contract_version'
           IS DISTINCT FROM 'harpy/iam-v1-adoption-plan/v1' OR
       (v_plan->>'tenant_id')::uuid IS DISTINCT FROM p_tenant_id OR
       v_plan->>'mapping_sha256' !~ '^[0-9a-f]{64}$' OR
       v_plan->>'source_sha256' !~ '^[0-9a-f]{64}$' OR
       pg_catalog.jsonb_typeof(v_plan->'categories') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_typeof(v_plan->'issues') IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption plan contract is invalid';
    END IF;
    IF pg_catalog.jsonb_array_length(v_plan->'categories') <> 15 OR
       (SELECT pg_catalog.array_agg(value->>'category' ORDER BY value->>'category')
        FROM pg_catalog.jsonb_array_elements(v_plan->'categories') AS value)
       IS DISTINCT FROM ARRAY[
           'actions', 'assignments', 'credentials', 'identifiers', 'invitations',
           'layouts', 'memberships', 'organizations', 'permission_requests',
           'permissions', 'portal', 'profiles', 'resources', 'roles', 'users'
       ]::text[] OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(v_plan->'categories') AS value
           WHERE (SELECT pg_catalog.array_agg(key ORDER BY key)
                  FROM pg_catalog.jsonb_object_keys(value) AS key)
                 IS DISTINCT FROM ARRAY['category','root_sha256','row_count']::text[]
              OR (value->>'row_count')::bigint < 0
              OR value->>'root_sha256' !~ '^[0-9a-f]{64}$'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption category evidence is invalid';
    END IF;
    v_issue_count := pg_catalog.jsonb_array_length(v_plan->'issues');
    v_eligible := (v_plan->>'eligible')::boolean;
    IF v_eligible IS DISTINCT FROM (v_issue_count = 0) OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(v_plan->'issues') AS value
        WHERE (SELECT pg_catalog.array_agg(key ORDER BY key)
               FROM pg_catalog.jsonb_object_keys(value) AS key)
              IS DISTINCT FROM ARRAY['category','code','source_key']::text[]
           OR value->>'category' = '' OR value->>'code' = ''
           OR pg_catalog.char_length(value->>'source_key') > 512
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption issue evidence is invalid';
    END IF;
    SELECT plan_sha256 INTO v_existing FROM iam_v1_adoption_runs
    WHERE operation_id = p_operation_id AND run_id = p_run_id FOR UPDATE;
    IF FOUND THEN
        IF v_existing IS DISTINCT FROM pg_catalog.decode(p_plan_sha256, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'IAM V1 adoption plan identity was reused';
        END IF;
        RETURN (SELECT state FROM iam_v1_adoption_runs
                WHERE operation_id = p_operation_id AND run_id = p_run_id);
    END IF;
    INSERT INTO iam_v1_adoption_runs (
        operation_id, run_id, tenant_id, contract_version, source_sha256,
        mapping_sha256, plan_sha256, state, eligible, issue_count, created_at, updated_at
    ) VALUES (
        p_operation_id, p_run_id, p_tenant_id, v_plan->>'contract_version',
        pg_catalog.decode(v_plan->>'source_sha256', 'hex'),
        pg_catalog.decode(v_plan->>'mapping_sha256', 'hex'),
        pg_catalog.decode(p_plan_sha256, 'hex'),
        CASE WHEN v_eligible THEN 'planned' ELSE 'blocked' END,
        v_eligible, v_issue_count, v_now, v_now
    );
    INSERT INTO iam_v1_adoption_category_evidence (
        operation_id, run_id, category, source_count, source_sha256, updated_at
    )
    SELECT p_operation_id, p_run_id, value->>'category',
        (value->>'row_count')::bigint,
        pg_catalog.decode(value->>'root_sha256', 'hex'), v_now
    FROM pg_catalog.jsonb_array_elements(v_plan->'categories') AS value;
    INSERT INTO iam_v1_adoption_issues (
        operation_id, run_id, ordinal, code, category, source_key, created_at
    )
    SELECT p_operation_id, p_run_id, position::integer, value->>'code',
        value->>'category', value->>'source_key', v_now
    FROM pg_catalog.jsonb_array_elements(v_plan->'issues')
        WITH ORDINALITY AS issue(value, position);
    RETURN CASE WHEN v_eligible THEN 'planned' ELSE 'blocked' END;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_plan_register_v1(
    name, uuid, uuid, uuid, bytea, text
) FROM PUBLIC;

