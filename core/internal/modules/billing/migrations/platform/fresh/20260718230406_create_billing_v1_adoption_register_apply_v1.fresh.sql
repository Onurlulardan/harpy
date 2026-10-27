-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_register_apply_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_plan_bytes bytea, p_last_key jsonb, p_batch_sha256 text,
    p_batch_ordinal integer, p_lease_manifest_sha256 text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_plan jsonb;
    v_evidence jsonb;
    v_evidence_bytes bytea;
    v_evidence_sha text;
    v_plan_sha text;
    v_authority_id uuid;
    v_lease_id uuid;
    v_expected_plan_sha text;
    v_cutover_generation bigint;
    v_state text;
    v_applied bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_batch_ordinal NOT BETWEEN 1 AND 2147483 OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR
       p_lease_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption registration principal denied';
    END IF;
    SELECT operation.lease_id, operation.plan_sha256,
        operation.cutover_generation
    INTO STRICT v_lease_id, v_expected_plan_sha, v_cutover_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id
      AND operation.run_id = p_run_id
      AND operation.scope_kind = p_scope_kind
      AND operation.scope_id = p_scope_id
      AND operation.domain = 'billing'
      AND operation.lease_manifest_sha256 = p_lease_manifest_sha256
      AND operation.state IN ('copying','shadow','quiescing')
    FOR SHARE;
    v_plan_sha := pg_catalog.encode(pg_catalog.sha256(p_plan_bytes), 'hex');
    v_plan := pg_catalog.convert_from(p_plan_bytes, 'UTF8')::jsonb;
    IF v_plan_sha IS DISTINCT FROM v_expected_plan_sha OR
       v_plan->>'scope_kind' IS DISTINCT FROM p_scope_kind OR
       (v_plan->>'scope_id')::uuid IS DISTINCT FROM p_scope_id OR
       v_plan->>'contract_version' IS DISTINCT FROM
           'forgecrud/billing-v1-adoption-plan/v1' OR
       (v_plan->>'eligible')::boolean IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption compiled plan differs';
    END IF;
    SELECT run.* INTO v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
    FOR UPDATE;
    IF FOUND THEN
        IF v_run.scope_kind IS DISTINCT FROM p_scope_kind OR
           v_run.scope_id IS DISTINCT FROM p_scope_id OR
           v_run.cutover_lease_id IS DISTINCT FROM v_lease_id OR
           v_run.cutover_lease_sha256 IS DISTINCT FROM
               pg_catalog.decode(p_lease_manifest_sha256, 'hex') OR
           v_run.plan_sha256 IS DISTINCT FROM pg_catalog.decode(v_plan_sha, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'billing V1 adoption registered run differs';
        END IF;
        v_applied := platform.billing_apply_v1_adoption_v1(
            p_operation_id, p_run_id
        );
        RETURN pg_catalog.encode(v_run.source_sha256, 'hex');
    END IF;
    IF p_scope_kind = 'global' THEN
        SELECT authority.authority_id INTO STRICT v_authority_id
        FROM platform.billing_global_authorities AS authority
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = authority.scope_kind
         AND group_row.scope_id = authority.scope_id
         AND group_row.group_id = authority.authority_group_id
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind = authority.scope_kind
         AND domain_row.scope_id = authority.scope_id
         AND domain_row.domain = authority.authority_domain
        WHERE authority.scope_id = p_scope_id
          AND authority.authority_group_generation = group_row.generation
          AND authority.authority_domain_generation = domain_row.generation
          AND domain_row.generation = v_cutover_generation;
    ELSE
        SELECT authority.authority_id INTO STRICT v_authority_id
        FROM platform.billing_tenant_authorities AS authority
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = authority.scope_kind
         AND group_row.scope_id = authority.scope_id
         AND group_row.group_id = authority.authority_group_id
        JOIN platform.domain_cutovers AS billing_row
          ON billing_row.scope_kind = authority.scope_kind
         AND billing_row.scope_id = authority.scope_id
         AND billing_row.domain = authority.billing_domain
        JOIN platform.domain_cutovers AS forms_row
          ON forms_row.scope_kind = authority.scope_kind
         AND forms_row.scope_id = authority.scope_id
         AND forms_row.domain = authority.forms_domain
        WHERE authority.tenant_id = p_scope_id
          AND authority.authority_group_generation = group_row.generation
          AND authority.billing_generation = billing_row.generation
          AND authority.forms_generation = forms_row.generation
          AND billing_row.generation = v_cutover_generation;
    END IF;
    v_evidence := pg_catalog.jsonb_build_object(
        'batch_sha256', p_batch_sha256,
        'contract_version', 'forgecrud/billing-v1-cutover-evidence/v1',
        'evidence_generation', p_batch_ordinal,
        'lease_id', v_lease_id,
        'lease_manifest_sha256', p_lease_manifest_sha256,
        'operation_id', p_operation_id,
        'plan_sha256', v_plan_sha,
        'run_id', p_run_id,
        'scope_id', p_scope_id,
        'scope_kind', p_scope_kind,
        'session_seal_receipt_id',
            (p_last_key->>'session_seal_receipt_id')::uuid,
        'session_seal_sha256', p_last_key->>'session_seal_sha256'
    );
    v_evidence_bytes := pg_catalog.convert_to(v_evidence::text, 'UTF8');
    v_evidence_sha := pg_catalog.encode(
        pg_catalog.sha256(v_evidence_bytes), 'hex'
    );
    v_state := platform.billing_register_v1_adoption_v1(
        p_operation_id, p_run_id, v_authority_id, p_plan_bytes, v_plan_sha,
        v_evidence_bytes, v_evidence_sha
    );
    IF v_state IS DISTINCT FROM 'planned' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing V1 adoption registration was not planned';
    END IF;
    v_applied := platform.billing_apply_v1_adoption_v1(
        p_operation_id, p_run_id
    );
    IF v_applied < 0 THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'billing V1 adoption apply count is invalid';
    END IF;
    RETURN v_plan->>'source_sha256';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_register_apply_v1(
    uuid, uuid, text, uuid, bytea, jsonb, text, integer, text
) FROM PUBLIC;
