-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_worker_capability_mint_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_resource_id uuid,
    p_actual_source_schema_id uuid, p_source_revision bigint,
    p_source_sha256 bytea, p_target_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
    v_source_version bigint;
    v_target_schema_id uuid;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) OR p_plan_id IS NULL OR p_page_number < 1 OR
       p_resource_id IS NULL OR p_actual_source_schema_id IS NULL OR
       p_source_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_target_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile worker capability input is invalid';
    END IF;
    SELECT plan.resource_kind, source.version, plan.target_schema_id,
           page.after_resource_id, page.through_resource_id
    INTO v_resource_kind, v_source_version, v_target_schema_id,
         v_after_resource_id, v_through_resource_id
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_backfill_pages AS page ON page.plan_id = plan.id
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.id = p_plan_id AND plan.state IN ('planned','running')
      AND page.page_number = p_page_number AND page.job_id = p_job_id
      AND page.payload_reference = p_payload_reference
      AND page.payload_sha256 = p_payload_sha256 AND page.state = 'queued'
    FOR UPDATE OF plan, page, source, target;
    IF v_resource_kind IS NULL OR
       (v_after_resource_id IS NOT NULL AND
        p_resource_id <= v_after_resource_id) OR
       p_resource_id > v_through_resource_id OR NOT EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS actual
           WHERE actual.id = p_actual_source_schema_id
             AND actual.resource_kind = v_resource_kind
             AND actual.state IN ('active','retired')
             AND actual.version <= v_source_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile worker capability is stale';
    END IF;
    INSERT INTO iam_profile_backfill_worker_capabilities (
        backend_pid, transaction_id, plan_id, page_number, resource_kind,
        resource_id, actual_source_schema_id, source_revision, source_sha256,
        target_schema_id, target_sha256, created_at
    ) VALUES (
        pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(),
        p_plan_id, p_page_number, v_resource_kind, p_resource_id,
        p_actual_source_schema_id, p_source_revision, p_source_sha256,
        v_target_schema_id, p_target_sha256,
        pg_catalog.transaction_timestamp()
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_worker_capability_mint_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,
    uuid,bigint,bytea,bytea
) FROM PUBLIC;
