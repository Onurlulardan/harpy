-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_stale_job_disposition_v1(
    p_job platform.jobs,p_root_id uuid,p_action_id uuid,p_result_job_id uuid,
    p_result_root_id uuid,p_result_payload_ref text,p_result_dedupe_key text,
    p_group_generation bigint,p_domain_generation bigint,p_action text,
    p_payload_canonical bytea,p_payload_sha256 text,p_reason text,p_ticket text,
    p_actor_id uuid,p_causation_id uuid,p_source_run_at text,
    p_result_run_at text,p_occurred_at text,p_now timestamptz
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_document jsonb;
    v_expected jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_root_id IS NULL OR p_action_id IS NULL OR
       p_root_id IN (p_action_id,p_job.id) OR p_action_id = p_job.id OR
       p_action NOT IN ('adopt','cancel') OR p_job.payload_ref IS NULL OR
       p_payload_canonical IS NULL OR
       pg_catalog.octet_length(p_payload_canonical) NOT BETWEEN 2 AND 65536 OR
       p_payload_sha256 !~ '^[0-9a-f]{64}$' OR
       p_payload_sha256 IS DISTINCT FROM pg_catalog.encode(
           pg_catalog.sha256(p_payload_canonical),'hex') OR
       p_reason IS NULL OR p_reason <> pg_catalog.btrim(p_reason) OR
       pg_catalog.octet_length(p_reason) NOT BETWEEN 1 AND 2048 OR
       p_ticket IS NULL OR p_ticket <> pg_catalog.btrim(p_ticket) OR
       pg_catalog.octet_length(p_ticket) NOT BETWEEN 1 AND 256 OR
       p_actor_id IS NULL OR p_causation_id IS NULL OR
       p_source_run_at IS NULL OR p_source_run_at !~ 'Z$' OR
       p_source_run_at::timestamptz <> p_job.run_at OR
       p_occurred_at IS NULL OR p_occurred_at !~ 'Z$' OR
       p_occurred_at::timestamptz <> p_now OR
       p_now IS DISTINCT FROM pg_catalog.transaction_timestamp() OR
       p_group_generation < p_job.authority_group_generation OR
       p_domain_generation < p_job.cutover_generation OR
       (p_group_generation = p_job.authority_group_generation AND
        p_domain_generation = p_job.cutover_generation) OR
       NOT ((p_action = 'adopt' AND p_result_job_id IS NOT NULL AND
             p_result_job_id <> p_job.id AND p_result_payload_ref IS NOT NULL AND
             p_result_job_id NOT IN (p_root_id,p_action_id) AND
             pg_catalog.octet_length(p_result_payload_ref) BETWEEN 1 AND 2048 AND
             p_result_dedupe_key IS NOT NULL AND
             pg_catalog.octet_length(p_result_dedupe_key) BETWEEN 1 AND 512 AND
             p_result_run_at IS NOT NULL AND p_result_run_at ~ 'Z$' AND
             p_result_run_at::timestamptz =
                 GREATEST(p_job.run_at,p_now)) OR
            (p_action = 'cancel' AND p_result_job_id IS NULL AND
             p_result_root_id IS NULL AND p_result_payload_ref IS NULL AND
             p_result_dedupe_key IS NULL AND p_result_run_at IS NULL)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5684',
            MESSAGE = 'stale job disposition evidence input is invalid';
    END IF;
    BEGIN
        v_document := pg_catalog.convert_from(p_payload_canonical,'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE = 'P5684',
            MESSAGE = 'stale job disposition evidence payload is invalid';
    END;
    v_expected := pg_catalog.jsonb_build_object(
        'contract_version','stale-job-disposition.v1',
        'evidence_root_id',p_root_id,'disposition_id',p_action_id,
        'source_job_id',p_job.id,'decision',p_action,
        'result_job_id',p_result_job_id,
        'kind',p_job.kind,'effect_class',p_job.effect_class,
        'maximum_attempts',p_job.max_attempts,'priority',p_job.priority,
        'old_authority',pg_catalog.jsonb_build_object(
            'scope_kind',p_job.scope_kind,'scope_id',p_job.scope_id,
            'domain',p_job.authority_domain,
            'cutover_generation',p_job.cutover_generation,
            'writer_origin',p_job.writer_origin,
            'group_id',p_job.authority_group_id,
            'group_generation',p_job.authority_group_generation),
        'current_authority',pg_catalog.jsonb_build_object(
            'scope_kind',p_job.scope_kind,'scope_id',p_job.scope_id,
            'domain',p_job.authority_domain,
            'cutover_generation',p_domain_generation,'writer_origin','v2',
            'group_id',p_job.authority_group_id,
            'group_generation',p_group_generation),
        'source_payload',pg_catalog.jsonb_build_object(
            'revision',p_job.payload_revision,'sha256',p_job.payload_sha256,
            'reference',p_job.payload_ref,'dedupe_key',p_job.dedupe_key,
            'effect_idempotency_key',p_job.effect_idempotency_key),
        'result_payload',CASE WHEN p_action = 'adopt' THEN
            pg_catalog.jsonb_build_object(
                'revision',p_job.payload_revision,'sha256',p_job.payload_sha256,
                'reference',p_result_payload_ref,'dedupe_key',p_result_dedupe_key,
                'effect_idempotency_key',p_job.effect_idempotency_key)
            ELSE 'null'::jsonb END,
        'source_run_at',p_source_run_at,'result_run_at',p_result_run_at,
        'reason',p_reason,'ticket_reference',p_ticket,
        'actor_id',p_actor_id,'causation_id',p_causation_id,
        'occurred_at',p_occurred_at
    );
    IF v_document IS DISTINCT FROM v_expected THEN
        RAISE EXCEPTION USING ERRCODE = 'P5684',
            MESSAGE = 'stale job disposition evidence contract is invalid';
    END IF;
    RETURN pg_catalog.jsonb_build_object(
        'contract_version','stale-job-disposition.v1','source_job_id',p_job.id,
        'decision',p_action,'result_job_id',p_result_job_id,
        'result_root_id',p_result_root_id,'kind',p_job.kind,
        'effect_class',p_job.effect_class,'maximum_attempts',p_job.max_attempts,
        'priority',p_job.priority,'source_payload_reference',p_job.payload_ref,
        'source_payload_sha256',p_job.payload_sha256,
        'source_dedupe_key',p_job.dedupe_key,
        'effect_idempotency_key',p_job.effect_idempotency_key,
        'result_payload_reference',p_result_payload_ref,
        'result_dedupe_key',p_result_dedupe_key,
        'source_run_at',p_source_run_at,'result_run_at',p_result_run_at,
        'current_group_generation',p_group_generation,
        'current_domain_generation',p_domain_generation
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_validate_stale_job_disposition_v1(
    platform.jobs,uuid,uuid,uuid,uuid,text,text,bigint,bigint,text,bytea,text,
    text,text,uuid,uuid,text,text,text,timestamptz
) FROM PUBLIC;
