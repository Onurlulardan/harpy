-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_mappings_read_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea, p_plan_id uuid
)
RETURNS TABLE (
    source_field_id uuid, target_field_id uuid,
    converter_kind text, semantic_sha256 bytea
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
       ) OR NOT EXISTS (
           SELECT 1 FROM iam_profile_backfill_pages AS page
           WHERE page.plan_id = p_plan_id AND page.job_id = p_job_id
             AND page.payload_reference = p_payload_reference
             AND page.payload_sha256 = p_payload_sha256
             AND page.state = 'queued'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile backfill mapping authority is invalid';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM iam_profile_backfill_mappings AS mapping
        WHERE mapping.plan_id = p_plan_id) > 250 THEN
        RAISE EXCEPTION USING ERRCODE = '54000',
            MESSAGE = 'IAM profile backfill mapping bound exceeded';
    END IF;
    RETURN QUERY
    SELECT mapping.source_field_id, mapping.target_field_id,
           mapping.converter_kind, mapping.semantic_sha256
    FROM iam_profile_backfill_mappings AS mapping
    WHERE mapping.plan_id = p_plan_id ORDER BY mapping.ordinal;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_mappings_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) FROM PUBLIC;
