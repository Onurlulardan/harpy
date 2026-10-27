-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_replay_schema_submission_v1(
    p_tenant_id uuid,
    p_resource_kind text,
    p_resource_id uuid,
    p_resource_revision bigint,
    p_plan_contract_version text,
    p_compiler_version text,
    p_plan jsonb,
    p_canonical_plan_bytes bytea,
    p_plan_sha256 text,
    p_canonical_steps jsonb
)
RETURNS TABLE(job_id uuid, generation bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_row platform.schema_job_steps%ROWTYPE;
    v_step jsonb;
    v_index integer;
    v_step_sha256 text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'schema submission replay role is invalid';
    END IF;
    SELECT candidate.* INTO v_job
    FROM platform.schema_jobs AS candidate
    WHERE candidate.tenant_id = p_tenant_id
      AND candidate.owner = 'forms'
      AND candidate.plan_sha256 = p_plan_sha256
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
       OR (
            v_job.canonical_plan_bytes IS NOT NULL
            AND v_job.canonical_plan_bytes IS DISTINCT FROM p_canonical_plan_bytes
       )
       OR v_job.step_count IS DISTINCT FROM pg_catalog.jsonb_array_length(p_canonical_steps)
       THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'schema submission replay evidence differs';
    END IF;
    FOR v_index IN 0..v_job.step_count - 1 LOOP
        v_step := platform.platform_migrator_validate_submission_step_v1(
            p_plan_sha256,
            p_resource_kind,
            pg_catalog.decode(p_canonical_steps->>v_index, 'hex'),
            v_index,
            p_plan->'allowed_operations'
        );
        SELECT candidate.* INTO v_row
        FROM platform.schema_job_steps AS candidate
        WHERE candidate.job_id = v_job.job_id
          AND candidate.ordinal = v_index;
        v_step_sha256 := pg_catalog.encode(
            pg_catalog.sha256(pg_catalog.decode(p_canonical_steps->>v_index, 'hex')), 'hex'
        );
        IF NOT FOUND
           OR v_row.plan_sha256 IS DISTINCT FROM p_plan_sha256
           OR v_row.operation_id IS DISTINCT FROM v_step->>'operation_id'
           OR v_row.transaction_class IS DISTINCT FROM v_step->>'transaction_class'
           OR v_row.executor_kind IS DISTINCT FROM v_step->>'executor_kind'
           OR v_row.identifier_mapping IS DISTINCT FROM v_step->'identifier_mapping'
           OR v_row.arguments IS DISTINCT FROM v_step->'arguments'
           OR v_row.step_sha256 IS DISTINCT FROM v_step_sha256
           OR (
                v_row.canonical_step_bytes IS NOT NULL
                AND v_row.canonical_step_bytes
                    IS DISTINCT FROM pg_catalog.decode(p_canonical_steps->>v_index, 'hex')
           )
           THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'schema submission replay step differs';
        END IF;
    END LOOP;
    job_id := v_job.job_id;
    generation := v_job.generation;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_replay_schema_submission_v1(
    uuid, text, uuid, bigint, text, text, jsonb, bytea, text, jsonb
) FROM PUBLIC;
