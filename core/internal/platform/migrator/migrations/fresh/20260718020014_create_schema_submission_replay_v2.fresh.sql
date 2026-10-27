-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_replay_schema_submission_v2(
    p_tenant_id uuid, p_resource_kind text, p_resource_id uuid,
    p_resource_revision bigint, p_plan_contract_version text,
    p_compiler_version text, p_plan jsonb, p_canonical_plan_bytes bytea,
    p_plan_sha256 text, p_canonical_steps jsonb, p_authority_id uuid,
    p_authority_group_id text, p_group_generation bigint,
    p_billing_generation bigint, p_forms_generation bigint
)
RETURNS TABLE(job_id uuid,generation bigint,job_state text,terminal boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_row platform.schema_job_steps%ROWTYPE;
    v_step jsonb;
    v_index integer;
    v_step_bytes bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='schema V2 replay authority is invalid';
    END IF;
    SELECT candidate.* INTO v_job
    FROM platform.schema_jobs candidate
    WHERE candidate.tenant_id=p_tenant_id AND candidate.owner='forms'
      AND candidate.plan_sha256=p_plan_sha256
    FOR SHARE;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    IF v_job.resource_kind IS DISTINCT FROM p_resource_kind
       OR v_job.resource_id IS DISTINCT FROM p_resource_id
       OR v_job.resource_revision IS DISTINCT FROM p_resource_revision
       OR v_job.plan_contract_version IS DISTINCT FROM p_plan_contract_version
       OR v_job.compiler_version IS DISTINCT FROM p_compiler_version
       OR v_job.canonical_plan IS DISTINCT FROM p_plan
       OR v_job.canonical_plan_bytes IS DISTINCT FROM p_canonical_plan_bytes
       OR v_job.step_count IS DISTINCT FROM pg_catalog.jsonb_array_length(p_canonical_steps)
       OR v_job.submission_authority_id IS DISTINCT FROM p_authority_id
       OR v_job.submission_authority_group_id IS DISTINCT FROM p_authority_group_id
       OR v_job.submission_group_generation IS DISTINCT FROM p_group_generation
       OR v_job.submission_billing_generation IS DISTINCT FROM p_billing_generation
       OR v_job.submission_forms_generation IS DISTINCT FROM p_forms_generation THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='schema V2 submission replay evidence differs';
    END IF;
    FOR v_index IN 0..v_job.step_count-1 LOOP
        v_step_bytes := pg_catalog.decode(p_canonical_steps->>v_index,'hex');
        v_step := platform.platform_migrator_validate_submission_step_v2(
            p_plan_sha256,p_resource_kind,v_step_bytes,v_index,
            p_plan->'allowed_operations'
        );
        SELECT candidate.* INTO v_row
        FROM platform.schema_job_steps candidate
        WHERE candidate.job_id=v_job.job_id AND candidate.ordinal=v_index;
        IF NOT FOUND
           OR v_row.plan_sha256 IS DISTINCT FROM p_plan_sha256
           OR v_row.operation_id IS DISTINCT FROM v_step->>'operation_id'
           OR v_row.transaction_class IS DISTINCT FROM v_step->>'transaction_class'
           OR v_row.executor_kind IS DISTINCT FROM v_step->>'executor_kind'
           OR v_row.identifier_mapping IS DISTINCT FROM v_step->'identifier_mapping'
           OR v_row.arguments IS DISTINCT FROM v_step->'arguments'
           OR v_row.step_sha256 IS DISTINCT FROM pg_catalog.encode(
                pg_catalog.sha256(v_step_bytes),'hex')
           OR v_row.canonical_step_bytes IS DISTINCT FROM v_step_bytes THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='schema V2 submission replay step differs';
        END IF;
    END LOOP;
    job_id := v_job.job_id;
    generation := v_job.generation;
    job_state := v_job.state;
    terminal := v_job.state IN ('succeeded','failed','cancelled');
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_migrator_replay_schema_submission_v2(
    uuid,text,uuid,bigint,text,text,jsonb,bytea,text,jsonb,
    uuid,text,bigint,bigint,bigint
) FROM PUBLIC;
