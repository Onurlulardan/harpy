-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_accept_v1_adoption_v1(
    p_run_id uuid,
    p_zero_inventory_sha256 text,
    p_disposition_sha256 text,
    p_accepted_at timestamptz,
    p_acceptance_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.async_v1_adoption_runs%ROWTYPE;
    v_zero_observed_at timestamptz;
    v_disposition_count bigint;
    v_disposition_sha256 text;
    v_binding_count bigint;
    v_adopted_count bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_cluster_admin' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5660',
            MESSAGE = 'V1 async acceptance requires cluster authority';
    END IF;
    SELECT * INTO v_run FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id FOR UPDATE;
    IF v_run.state OPERATOR(pg_catalog.=) 'accepted' AND
       v_run.acceptance_sha256 OPERATOR(pg_catalog.=) p_acceptance_sha256 THEN
        RETURN;
    END IF;
    IF NOT FOUND OR v_run.state IS DISTINCT FROM 'disposing' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5661',
            MESSAGE = 'V1 async adoption is not ready for acceptance';
    END IF;
    SELECT observed_at INTO v_zero_observed_at
    FROM platform.async_v1_zero_inventory_receipts
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
      AND inventory_sha256 OPERATOR(pg_catalog.=) v_run.inventory_sha256
      AND receipt_sha256 OPERATOR(pg_catalog.=) p_zero_inventory_sha256
    FOR SHARE;
    IF NOT FOUND OR p_accepted_at < v_zero_observed_at OR
       p_accepted_at > pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P5662',
            MESSAGE = 'V1 zero inventory receipt is missing or stale';
    END IF;
    SELECT pg_catalog.count(*),
           pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
               COALESCE(pg_catalog.string_agg(
                   receipt_sha256, E'\n' ORDER BY receipt_sha256
               ), ''), 'UTF8'
           )), 'hex')
    INTO v_disposition_count, v_disposition_sha256
    FROM platform.async_v1_adoption_dispositions
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    IF v_disposition_count IS DISTINCT FROM v_run.item_count OR
       v_disposition_sha256 IS DISTINCT FROM p_disposition_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5663',
            MESSAGE = 'V1 disposition count or checksum differs';
    END IF;
    SELECT pg_catalog.count(*) INTO v_binding_count
    FROM platform.async_v1_adoption_job_bindings
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    SELECT pg_catalog.count(*) INTO v_adopted_count
    FROM platform.async_v1_adoption_dispositions
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
      AND disposition OPERATOR(pg_catalog.=) 'adopted';
    IF v_binding_count IS DISTINCT FROM v_adopted_count OR EXISTS (
        SELECT 1
        FROM platform.async_v1_adoption_dispositions AS disposition
        JOIN platform.async_v1_adoption_items AS item USING (
            run_id, inventory_class, source_name, source_item_id
        )
        LEFT JOIN platform.async_v1_adoption_job_bindings AS binding USING (
            run_id, inventory_class, source_name, source_item_id
        )
        LEFT JOIN platform.jobs AS job ON job.id OPERATOR(pg_catalog.=) binding.job_id
        WHERE disposition.run_id OPERATOR(pg_catalog.=) p_run_id
          AND disposition.disposition OPERATOR(pg_catalog.=) 'adopted'
          AND (binding.job_id IS NULL OR job.id IS NULL OR
               disposition.target_job_id IS DISTINCT FROM binding.job_id OR
               job.writer_origin IS DISTINCT FROM 'v1' OR
               job.state IS DISTINCT FROM 'queued' OR
               job.cutover_generation IS DISTINCT FROM v_run.target_generation OR
               job.authority_group_generation IS DISTINCT FROM
                   v_run.target_group_generation OR
               job.payload_sha256 IS DISTINCT FROM item.target_payload_sha256 OR
               job.dedupe_key IS DISTINCT FROM item.idempotency_identity)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5664',
            MESSAGE = 'V1 adopted job binding is incomplete';
    END IF;
    UPDATE platform.async_v1_adoption_runs
    SET state = 'accepted', accepted_at = p_accepted_at,
        acceptance_sha256 = p_acceptance_sha256
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_accept_v1_adoption_v1(
    uuid,text,text,timestamptz,text
) FROM PUBLIC;
