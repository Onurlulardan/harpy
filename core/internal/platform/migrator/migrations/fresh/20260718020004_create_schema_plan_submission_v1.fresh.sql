-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_submit_schema_plan_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_job_id uuid,
    p_resource_kind text,
    p_resource_id uuid,
    p_resource_revision bigint,
    p_plan_contract_version text,
    p_compiler_version text,
    p_postgresql_major integer,
    p_canonical_plan_bytes bytea,
    p_plan_sha256 text,
    p_canonical_steps jsonb
)
RETURNS TABLE(
    job_id uuid,
    generation bigint,
    plan_sha256 text,
    replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_plan jsonb;
    v_step jsonb;
    v_ordinal integer;
    v_existing_job_id uuid;
    v_existing_generation bigint;
BEGIN
    v_plan := platform.platform_migrator_assert_schema_submission_v1(
        p_scope_schema,
        p_tenant_id,
        p_job_id,
        p_resource_kind,
        p_resource_id,
        p_resource_revision,
        p_plan_contract_version,
        p_compiler_version,
        p_postgresql_major,
        p_canonical_plan_bytes,
        p_plan_sha256,
        p_canonical_steps
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'forgecrud:v2:schema-revision:' || p_tenant_id::text || ':forms:' ||
        p_resource_kind || ':' || p_resource_id::text,
        0
    ));
    SELECT existing.job_id, existing.generation
    INTO v_existing_job_id, v_existing_generation
    FROM platform.platform_migrator_replay_schema_submission_v1(
        p_tenant_id,
        p_resource_kind,
        p_resource_id,
        p_resource_revision,
        p_plan_contract_version,
        p_compiler_version,
        v_plan,
        p_canonical_plan_bytes,
        p_plan_sha256,
        p_canonical_steps
    ) AS existing;
    IF FOUND THEN
        job_id := v_existing_job_id;
        generation := v_existing_generation;
        plan_sha256 := p_plan_sha256;
        replay := TRUE;
        RETURN NEXT;
        RETURN;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.schema_jobs AS existing
        WHERE existing.job_id = p_job_id
        FOR SHARE
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'schema submission job identity is already owned';
    END IF;
    INSERT INTO platform.schema_jobs (
        job_id,
        tenant_id,
        owner,
        resource_kind,
        resource_id,
        resource_revision,
        compiler_version,
        plan_contract_version,
        canonical_plan,
        canonical_plan_bytes,
        plan_sha256,
        step_count,
        available_at
    ) VALUES (
        p_job_id,
        p_tenant_id,
        'forms',
        p_resource_kind,
        p_resource_id,
        p_resource_revision,
        p_compiler_version,
        p_plan_contract_version,
        v_plan,
        p_canonical_plan_bytes,
        p_plan_sha256,
        pg_catalog.jsonb_array_length(p_canonical_steps),
        pg_catalog.statement_timestamp()
    );
    FOR v_ordinal IN 0..pg_catalog.jsonb_array_length(p_canonical_steps) - 1 LOOP
        v_step := platform.platform_migrator_validate_submission_step_v1(
            p_plan_sha256,
            p_resource_kind,
            pg_catalog.decode(p_canonical_steps->>v_ordinal, 'hex'),
            v_ordinal,
            v_plan->'allowed_operations'
        );
        IF v_plan->'steps'->v_ordinal IS DISTINCT FROM v_step THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'schema submission plan and step differ';
        END IF;
        INSERT INTO platform.schema_job_steps (
            job_id,
            ordinal,
            plan_sha256,
            operation_id,
            transaction_class,
            executor_kind,
            identifier_mapping,
            arguments,
            canonical_step_bytes,
            step_sha256,
            next_executor
        ) VALUES (
            p_job_id,
            v_ordinal,
            p_plan_sha256,
            v_step->>'operation_id',
            v_step->>'transaction_class',
            v_step->>'executor_kind',
            v_step->'identifier_mapping',
            v_step->'arguments',
            pg_catalog.decode(p_canonical_steps->>v_ordinal, 'hex'),
            pg_catalog.encode(
                pg_catalog.sha256(pg_catalog.decode(p_canonical_steps->>v_ordinal, 'hex')), 'hex'
            ),
            v_step->>'executor_kind'
        );
    END LOOP;
    job_id := p_job_id;
    generation := 1;
    plan_sha256 := p_plan_sha256;
    replay := FALSE;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_submit_schema_plan_v1(
    name, uuid, uuid, text, uuid, bigint, text, text, integer, bytea, text, jsonb
) FROM PUBLIC;
