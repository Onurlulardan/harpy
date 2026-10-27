-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_apply_v3(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_resource_id uuid,
    p_actual_schema_version_id uuid, p_expected_revision bigint,
    p_source_sha256 bytea, p_target_document bytea, p_target_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_profile_worker_capability_mint_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        p_payload_revision, p_payload_reference, p_payload_sha256,
        p_plan_id, p_page_number, p_resource_id,
        p_actual_schema_version_id, p_expected_revision,
        p_source_sha256, p_target_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile worker capability was denied';
    END IF;
    RETURN iam_maintenance_profile_backfill_apply_v2(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        p_payload_revision, p_payload_reference, p_payload_sha256,
        p_plan_id, p_page_number, p_resource_id,
        p_actual_schema_version_id, p_expected_revision,
        p_source_sha256, p_target_document, p_target_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_apply_v3(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) FROM PUBLIC;
