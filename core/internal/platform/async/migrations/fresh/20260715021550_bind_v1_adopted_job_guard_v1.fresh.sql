-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_v1_adopted_job_v1(
    p_job platform.jobs
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       p_job.writer_origin IS DISTINCT FROM 'v1' OR
       p_job.state IS DISTINCT FROM 'queued' OR
       p_job.attempts IS DISTINCT FROM 0 OR p_job.lease_generation IS DISTINCT FROM 0 OR
       p_job.lease_owner_id IS NOT NULL OR p_job.lease_until IS NOT NULL OR
       p_job.parked_at IS NOT NULL OR p_job.effect_started_at IS NOT NULL OR
       p_job.completed_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5644',
            MESSAGE = 'V1 adopted job initial state is invalid';
    END IF;
    SELECT run.run_id INTO v_run_id
    FROM platform.async_v1_adoption_job_bindings AS binding
    JOIN platform.async_v1_adoption_items AS item USING (
        run_id, inventory_class, source_name, source_item_id
    )
    JOIN platform.async_v1_adoption_runs AS run USING (run_id)
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind OPERATOR(pg_catalog.=) run.scope_kind
     AND group_row.scope_id OPERATOR(pg_catalog.=) run.scope_id
     AND group_row.group_id OPERATOR(pg_catalog.=) run.group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind OPERATOR(pg_catalog.=) run.scope_kind
     AND domain_row.scope_id OPERATOR(pg_catalog.=) run.scope_id
     AND domain_row.domain OPERATOR(pg_catalog.=) run.domain
    WHERE binding.job_id OPERATOR(pg_catalog.=) p_job.id
      AND binding.item_sha256 OPERATOR(pg_catalog.=) item.item_sha256
      AND run.state OPERATOR(pg_catalog.=) 'disposing'
      AND group_row.state OPERATOR(pg_catalog.=) 'quiescing'
      AND group_row.writer OPERATOR(pg_catalog.=) 'none'
      AND group_row.generation OPERATOR(pg_catalog.=) run.source_group_generation
      AND domain_row.state OPERATOR(pg_catalog.=) 'quiescing'
      AND domain_row.writer OPERATOR(pg_catalog.=) 'none'
      AND domain_row.generation OPERATOR(pg_catalog.=) run.source_generation
      AND p_job.scope_kind OPERATOR(pg_catalog.=) run.scope_kind
      AND p_job.scope_id OPERATOR(pg_catalog.=) run.scope_id
      AND p_job.authority_group_id OPERATOR(pg_catalog.=) run.group_id
      AND p_job.authority_group_generation OPERATOR(pg_catalog.=) run.target_group_generation
      AND p_job.authority_domain OPERATOR(pg_catalog.=) run.domain
      AND p_job.cutover_generation OPERATOR(pg_catalog.=) run.target_generation
      AND p_job.kind OPERATOR(pg_catalog.=) item.target_kind
      AND p_job.payload_revision OPERATOR(pg_catalog.=) item.target_payload_revision
      AND p_job.payload_ref OPERATOR(pg_catalog.=) item.target_payload_ref
      AND p_job.payload_sha256 OPERATOR(pg_catalog.=) item.target_payload_sha256
      AND p_job.dedupe_key OPERATOR(pg_catalog.=) item.idempotency_identity
      AND p_job.effect_class OPERATOR(pg_catalog.=) item.target_effect_class
      AND p_job.effect_idempotency_key IS NOT DISTINCT FROM
          item.target_effect_idempotency_key
      AND p_job.priority OPERATOR(pg_catalog.=) item.target_priority
      AND p_job.run_at OPERATOR(pg_catalog.=) item.target_run_at
      AND p_job.max_attempts OPERATOR(pg_catalog.=) item.target_max_attempts
      AND ((run.scope_kind OPERATOR(pg_catalog.=) 'tenant' AND
            p_job.tenant_id OPERATOR(pg_catalog.=) run.scope_id) OR
           (run.scope_kind OPERATOR(pg_catalog.=) 'global' AND p_job.tenant_id IS NULL))
      AND NOT EXISTS (
          SELECT 1 FROM platform.async_v1_adoption_dispositions AS disposition
          WHERE disposition.run_id OPERATOR(pg_catalog.=) run.run_id
            AND disposition.inventory_class OPERATOR(pg_catalog.=) item.inventory_class
            AND disposition.source_name OPERATOR(pg_catalog.=) item.source_name
            AND disposition.source_item_id OPERATOR(pg_catalog.=) item.source_item_id
      )
    FOR SHARE OF binding, item, run, group_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5645',
            MESSAGE = 'V1 adopted job has no exact quiesced inventory binding';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_assert_v1_adopted_job_v1(
    platform.jobs
) FROM PUBLIC;
