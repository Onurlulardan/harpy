-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_billing_forms_rehearsal_sha256_v1(
    p_tenant_id uuid,
    p_generation bigint,
    p_billing_operation_id uuid,
    p_billing_run_id uuid,
    p_billing_rollback_receipt_id uuid,
    p_billing_source_count bigint,
    p_billing_target_count bigint,
    p_billing_acceptance_sha256 text,
    p_billing_rollback_sha256 text,
    p_forms_operation_id uuid,
    p_forms_run_id uuid,
    p_forms_rollback_receipt_id uuid,
    p_forms_source_count bigint,
    p_forms_target_count bigint,
    p_forms_acceptance_sha256 text,
    p_forms_rollback_sha256 text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN pg_catalog.encode(pg_catalog.sha256(
    pg_catalog.convert_to(
        'harpy/cutover/billing-forms-rehearsal/v2', 'UTF8') ||
    pg_catalog.decode('00','hex') || pg_catalog.uuid_send(p_tenant_id) ||
    pg_catalog.int8send(p_generation) ||
    pg_catalog.convert_to('billing','UTF8') ||
    pg_catalog.uuid_send(p_billing_operation_id) ||
    pg_catalog.uuid_send(p_billing_run_id) ||
    pg_catalog.uuid_send(p_billing_rollback_receipt_id) ||
    pg_catalog.int8send(p_billing_source_count) ||
    pg_catalog.int8send(p_billing_target_count) ||
    pg_catalog.int8send(0::bigint) ||
    pg_catalog.decode(p_billing_acceptance_sha256,'hex') ||
    pg_catalog.decode(p_billing_rollback_sha256,'hex') ||
    pg_catalog.convert_to('forms','UTF8') ||
    pg_catalog.uuid_send(p_forms_operation_id) ||
    pg_catalog.uuid_send(p_forms_run_id) ||
    pg_catalog.uuid_send(p_forms_rollback_receipt_id) ||
    pg_catalog.int8send(p_forms_source_count) ||
    pg_catalog.int8send(p_forms_target_count) ||
    pg_catalog.int8send(0::bigint) ||
    pg_catalog.decode(p_forms_acceptance_sha256,'hex') ||
    pg_catalog.decode(p_forms_rollback_sha256,'hex') ||
    pg_catalog.decode('010100','hex')
),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_billing_forms_rehearsal_sha256_v1(
    uuid,bigint,uuid,uuid,uuid,bigint,bigint,text,text,
    uuid,uuid,uuid,bigint,bigint,text,text
) FROM PUBLIC;
