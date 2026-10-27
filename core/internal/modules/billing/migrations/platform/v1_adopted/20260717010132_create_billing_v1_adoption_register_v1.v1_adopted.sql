-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_register_v1_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_authority_id uuid,
    p_plan_bytes bytea,
    p_plan_sha256 text,
    p_cutover_evidence_bytes bytea,
    p_cutover_envelope_sha256 text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_plan jsonb;
    v_evidence jsonb;
    v_scope_kind text;
    v_scope_id uuid;
    v_source_sha256 text;
    v_eligible boolean;
    v_issue_count integer;
    v_existing bytea;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_operation_id IS NULL OR p_run_id IS NULL
       OR p_authority_id IS NULL
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_cutover_envelope_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(p_plan_bytes) NOT BETWEEN 2 AND 67108864
       OR pg_catalog.octet_length(p_cutover_evidence_bytes) NOT BETWEEN 2 AND 16384
       OR pg_catalog.encode(pg_catalog.sha256(p_plan_bytes), 'hex')
            IS DISTINCT FROM p_plan_sha256
       OR pg_catalog.encode(pg_catalog.sha256(p_cutover_evidence_bytes), 'hex')
            IS DISTINCT FROM p_cutover_envelope_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption registration is denied';
    END IF;
    v_plan := pg_catalog.convert_from(p_plan_bytes, 'UTF8')::jsonb;
    v_evidence := pg_catalog.convert_from(p_cutover_evidence_bytes, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_plan) IS DISTINCT FROM 'object'
       OR (SELECT pg_catalog.array_agg(key ORDER BY key)
           FROM pg_catalog.jsonb_object_keys(v_plan) AS key)
          IS DISTINCT FROM ARRAY[
              'base_assignments', 'campaigns', 'categories',
              'contract_version', 'discounts', 'eligible',
              'gauge_openings', 'issues', 'legacy_events',
              'manual_provider_account_id',
              'offerings', 'redemptions', 'scope_id', 'scope_kind',
              'snapshot_at', 'source_sha256', 'source_timezone',
              'subscriptions', 'usage_openings'
          ]::text[]
       OR v_plan->>'contract_version'
            IS DISTINCT FROM 'harpy/billing-v1-adoption-plan/v1'
       OR v_plan->>'source_timezone' IS DISTINCT FROM 'UTC'
       OR v_plan->>'source_sha256' !~ '^[0-9a-f]{64}$'
       OR pg_catalog.jsonb_typeof(v_plan->'categories') IS DISTINCT FROM 'array'
       OR pg_catalog.jsonb_typeof(v_plan->'issues') IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption plan contract is invalid';
    END IF;
    v_scope_kind := v_plan->>'scope_kind';
    v_scope_id := (v_plan->>'scope_id')::uuid;
    v_source_sha256 := v_plan->>'source_sha256';
    v_eligible := (v_plan->>'eligible')::boolean;
    v_issue_count := pg_catalog.jsonb_array_length(v_plan->'issues');
    IF pg_catalog.jsonb_typeof(v_evidence) IS DISTINCT FROM 'object'
       OR (SELECT pg_catalog.array_agg(key ORDER BY key)
           FROM pg_catalog.jsonb_object_keys(v_evidence) AS key)
          IS DISTINCT FROM ARRAY[
              'batch_sha256', 'contract_version', 'evidence_generation',
              'lease_id', 'lease_manifest_sha256', 'operation_id',
              'plan_sha256', 'run_id', 'scope_id', 'scope_kind',
              'session_seal_receipt_id', 'session_seal_sha256'
          ]::text[]
       OR v_evidence->>'contract_version' IS DISTINCT FROM
            'harpy/billing-v1-cutover-evidence/v1'
       OR (v_evidence->>'operation_id')::uuid IS DISTINCT FROM p_operation_id
       OR (v_evidence->>'run_id')::uuid IS DISTINCT FROM p_run_id
       OR v_evidence->>'scope_kind' IS DISTINCT FROM v_scope_kind
       OR (v_evidence->>'scope_id')::uuid IS DISTINCT FROM v_scope_id
       OR v_evidence->>'plan_sha256' IS DISTINCT FROM p_plan_sha256
       OR (v_evidence->>'evidence_generation')::bigint < 1
       OR v_evidence->>'lease_manifest_sha256' !~ '^[0-9a-f]{64}$'
       OR v_evidence->>'batch_sha256' !~ '^[0-9a-f]{64}$'
       OR v_evidence->>'session_seal_sha256' !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing V1 verified cutover envelope is invalid';
    END IF;
    IF v_scope_kind NOT IN ('global', 'tenant')
       OR v_scope_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR v_eligible IS DISTINCT FROM (v_issue_count = 0)
       OR EXISTS (
           SELECT 1
           FROM pg_catalog.jsonb_array_elements(v_plan->'categories') AS value
           WHERE (SELECT pg_catalog.array_agg(key ORDER BY key)
                  FROM pg_catalog.jsonb_object_keys(value) AS key)
                 IS DISTINCT FROM
                    ARRAY['category','root_sha256','row_count']::text[]
              OR value->>'root_sha256' !~ '^[0-9a-f]{64}$'
              OR (value->>'row_count')::bigint < 0
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption plan evidence is invalid';
    END IF;
    SELECT plan_sha256 INTO v_existing
    FROM platform.billing_v1_adoption_runs
    WHERE operation_id = p_operation_id AND run_id = p_run_id
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing IS DISTINCT FROM pg_catalog.decode(p_plan_sha256, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'billing V1 adoption operation identity was reused';
        END IF;
        RETURN (
            SELECT state FROM platform.billing_v1_adoption_runs
            WHERE operation_id = p_operation_id AND run_id = p_run_id
        );
    END IF;
    INSERT INTO platform.billing_v1_adoption_runs (
        operation_id, run_id, scope_kind, scope_id, authority_id,
        cutover_lease_id, cutover_evidence_generation,
        cutover_lease_sha256, cutover_batch_sha256,
        cutover_session_seal_receipt_id, cutover_session_seal_sha256,
        cutover_envelope_sha256,
        contract_version, plan_sha256, source_sha256, plan_bytes,
        plan_document, state, eligible, issue_count, snapshot_at,
        source_timezone, created_at, updated_at
    ) VALUES (
        p_operation_id, p_run_id, v_scope_kind, v_scope_id, p_authority_id,
        (v_evidence->>'lease_id')::uuid,
        (v_evidence->>'evidence_generation')::bigint,
        pg_catalog.decode(v_evidence->>'lease_manifest_sha256', 'hex'),
        pg_catalog.decode(v_evidence->>'batch_sha256', 'hex'),
        (v_evidence->>'session_seal_receipt_id')::uuid,
        pg_catalog.decode(v_evidence->>'session_seal_sha256', 'hex'),
        pg_catalog.decode(p_cutover_envelope_sha256, 'hex'),
        v_plan->>'contract_version', pg_catalog.decode(p_plan_sha256, 'hex'),
        pg_catalog.decode(v_source_sha256, 'hex'), p_plan_bytes, v_plan,
        CASE WHEN v_eligible THEN 'planned' ELSE 'blocked' END,
        v_eligible, v_issue_count, (v_plan->>'snapshot_at')::timestamptz,
        'UTC', v_now, v_now
    );
    INSERT INTO platform.billing_v1_adoption_categories (
        operation_id, run_id, scope_kind, scope_id, authority_id,
        category, source_count, source_sha256, updated_at
    )
    SELECT p_operation_id, p_run_id, v_scope_kind, v_scope_id,
        p_authority_id, value->>'category', (value->>'row_count')::bigint,
        pg_catalog.decode(value->>'root_sha256', 'hex'), v_now
    FROM pg_catalog.jsonb_array_elements(v_plan->'categories') AS value;
    RETURN CASE WHEN v_eligible THEN 'planned' ELSE 'blocked' END;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_register_v1_adoption_v1(
    uuid, uuid, uuid, bytea, text, bytea, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_register_v1_adoption_v1(
    uuid, uuid, uuid, bytea, text, bytea, text
) TO harpy_adopter;
-- +goose StatementEnd
