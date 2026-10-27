-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_page_read_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea
)
RETURNS TABLE (
    plan_id uuid, resource_kind text, source_schema_id uuid,
    target_schema_id uuid, converter_version text, plan_state text,
    page_number integer, after_resource_id uuid, through_resource_id uuid,
    expected_count integer, source_sha256 bytea,
    expected_target_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile backfill page authority is invalid';
    END IF;
    RETURN QUERY
    SELECT plan.id, plan.resource_kind, plan.source_schema_id,
           plan.target_schema_id, plan.converter_version, plan.state,
           page.page_number, page.after_resource_id, page.through_resource_id,
           page.expected_count, page.source_sha256,
           page.expected_target_sha256
    FROM iam_profile_backfill_pages AS page
    JOIN iam_profile_backfill_plans AS plan ON plan.id = page.plan_id
    WHERE page.job_id = p_job_id
      AND page.payload_reference = p_payload_reference
      AND page.payload_sha256 = p_payload_sha256
      AND page.state = 'queued' AND plan.state IN ('planned', 'running');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_page_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) FROM PUBLIC;
