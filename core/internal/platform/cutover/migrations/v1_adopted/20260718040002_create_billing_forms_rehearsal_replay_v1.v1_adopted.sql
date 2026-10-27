-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_billing_forms_rehearsal_replay_v1(
    p_receipt_id uuid,
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
    p_forms_rollback_sha256 text,
    p_group_sha256 text,
    p_recorded_at timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN (SELECT pg_catalog.encode(receipt.group_sha256,'hex')
FROM platform.cutover_billing_forms_rehearsals AS receipt
WHERE receipt.receipt_id=p_receipt_id
  AND receipt.tenant_id=p_tenant_id
  AND receipt.group_generation=p_generation
  AND receipt.billing_operation_id=p_billing_operation_id
  AND receipt.billing_run_id=p_billing_run_id
  AND receipt.billing_rollback_receipt_id=p_billing_rollback_receipt_id
  AND receipt.billing_source_count=p_billing_source_count
  AND receipt.billing_target_count=p_billing_target_count
  AND receipt.billing_acceptance_sha256=
      pg_catalog.decode(p_billing_acceptance_sha256,'hex')
  AND receipt.billing_rollback_sha256=
      pg_catalog.decode(p_billing_rollback_sha256,'hex')
  AND receipt.forms_operation_id=p_forms_operation_id
  AND receipt.forms_run_id=p_forms_run_id
  AND receipt.forms_rollback_receipt_id=p_forms_rollback_receipt_id
  AND receipt.forms_source_count=p_forms_source_count
  AND receipt.forms_target_count=p_forms_target_count
  AND receipt.forms_acceptance_sha256=
      pg_catalog.decode(p_forms_acceptance_sha256,'hex')
  AND receipt.forms_rollback_sha256=
      pg_catalog.decode(p_forms_rollback_sha256,'hex')
  AND receipt.group_sha256=pg_catalog.decode(p_group_sha256,'hex')
  AND receipt.recorded_at=p_recorded_at);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_billing_forms_rehearsal_replay_v1(
    uuid,uuid,bigint,uuid,uuid,uuid,bigint,bigint,text,text,
    uuid,uuid,uuid,bigint,bigint,text,text,text,timestamptz
) FROM PUBLIC;
