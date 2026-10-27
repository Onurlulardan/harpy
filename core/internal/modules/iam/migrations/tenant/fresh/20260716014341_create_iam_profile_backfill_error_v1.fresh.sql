-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_error_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_resource_id uuid,
    p_error_code text, p_evidence_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_maintenance_claim_assert_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.profile.backfill', p_payload_revision,
        p_payload_reference, p_payload_sha256
    ) OR p_plan_id IS NULL OR p_page_number < 1 OR p_resource_id IS NULL OR
       p_error_code NOT IN (
           'converter_missing', 'source_changed', 'document_invalid',
           'target_schema_invalid', 'checksum_mismatch'
       ) OR pg_catalog.octet_length(p_evidence_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (
           SELECT 1 FROM iam_profile_backfill_pages AS page
           WHERE page.plan_id = p_plan_id AND page.page_number = p_page_number
             AND page.job_id = p_job_id AND page.state = 'queued'
             AND (page.after_resource_id IS NULL OR
                  p_resource_id > page.after_resource_id)
             AND p_resource_id <= page.through_resource_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill error evidence is invalid';
    END IF;
    INSERT INTO iam_profile_backfill_errors (
        plan_id, page_number, resource_id, error_code,
        evidence_sha256, created_at
    ) VALUES (
        p_plan_id, p_page_number, p_resource_id, p_error_code,
        p_evidence_sha256, pg_catalog.transaction_timestamp()
    ) ON CONFLICT (plan_id, page_number, resource_id) DO UPDATE SET
        error_code = EXCLUDED.error_code,
        evidence_sha256 = EXCLUDED.evidence_sha256;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_error_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,text,bytea
) FROM PUBLIC;
