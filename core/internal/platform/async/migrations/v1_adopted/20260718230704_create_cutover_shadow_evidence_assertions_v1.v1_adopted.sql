-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_cutover_effect_receipt_v1(
    p_receipt platform.idempotency_receipts
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
BEGIN
    SELECT job.* INTO v_job FROM platform.jobs AS job
    JOIN platform.cutover_job_bindings AS binding ON binding.job_id=job.id
    WHERE job.id=p_receipt.causation_id FOR SHARE OF job,binding;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'cutover effect receipt job is missing';
    END IF;
    PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
    IF p_receipt.record_kind IS DISTINCT FROM 'effect' OR
       p_receipt.tenant_id IS DISTINCT FROM v_job.tenant_id OR
       p_receipt.scope_kind IS DISTINCT FROM v_job.scope_kind OR
       p_receipt.scope_id IS DISTINCT FROM v_job.scope_id OR
       p_receipt.authority_group_id IS DISTINCT FROM v_job.authority_group_id OR
       p_receipt.authority_group_generation IS DISTINCT FROM
           v_job.authority_group_generation OR
       p_receipt.authority_domain IS DISTINCT FROM v_job.authority_domain OR
       p_receipt.cutover_generation IS DISTINCT FROM v_job.cutover_generation OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.effect_kind IS DISTINCT FROM v_job.kind OR
       p_receipt.effect_revision IS DISTINCT FROM v_job.payload_revision OR
       p_receipt.effect_key IS DISTINCT FROM v_job.effect_idempotency_key OR
       p_receipt.provider IS DISTINCT FROM 'documents.adoption' OR
       p_receipt.operation IS DISTINCT FROM 'copy' OR NOT EXISTS (
           SELECT 1 FROM platform.job_attempts AS attempt
           WHERE attempt.job_id=v_job.id AND attempt.finished_at IS NULL
             AND attempt.effect_id=p_receipt.effect_id
             AND attempt.effect_request_sha256=p_receipt.request_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'cutover effect receipt differs';
    END IF;
END;
$function$;
CREATE FUNCTION platform.platform_async_assert_cutover_dead_letter_v1(
    p_letter platform.dead_letters
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
BEGIN
    SELECT job.* INTO v_job FROM platform.jobs AS job
    JOIN platform.cutover_job_bindings AS binding ON binding.job_id=job.id
    WHERE job.id=p_letter.source_id FOR SHARE OF job,binding;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'cutover dead letter job is missing';
    END IF;
    PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
    IF p_letter.record_kind IS DISTINCT FROM 'root' OR
       p_letter.source_kind IS DISTINCT FROM 'job' OR
       p_letter.tenant_id IS DISTINCT FROM v_job.tenant_id OR
       p_letter.scope_kind IS DISTINCT FROM v_job.scope_kind OR
       p_letter.scope_id IS DISTINCT FROM v_job.scope_id OR
       p_letter.authority_group_id IS DISTINCT FROM v_job.authority_group_id OR
       p_letter.authority_group_generation IS DISTINCT FROM
           v_job.authority_group_generation OR
       p_letter.authority_domain IS DISTINCT FROM v_job.authority_domain OR
       p_letter.cutover_generation IS DISTINCT FROM v_job.cutover_generation OR
       p_letter.writer_origin IS DISTINCT FROM 'v2' OR
       p_letter.source_contract IS DISTINCT FROM v_job.kind OR
       p_letter.source_version IS DISTINCT FROM v_job.payload_revision OR
       p_letter.source_sha256 IS DISTINCT FROM v_job.payload_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'cutover dead letter differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_assert_cutover_effect_receipt_v1(
    platform.idempotency_receipts) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_assert_cutover_dead_letter_v1(
    platform.dead_letters) FROM PUBLIC;
