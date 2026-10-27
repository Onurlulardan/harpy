-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_cutover_guard_v2_async_adoption_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run_id uuid;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'domain_cutovers' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5665',
            MESSAGE = 'async V2 switch guard target is invalid';
    END IF;
    IF NOT (OLD.state IS DISTINCT FROM 'v2' AND
            NEW.state OPERATOR(pg_catalog.=) 'v2') THEN
        RETURN NEW;
    END IF;
    IF OLD.state IS DISTINCT FROM 'quiescing' OR OLD.writer IS DISTINCT FROM 'none' OR
       NEW.writer IS DISTINCT FROM 'v2' OR
       NEW.generation IS DISTINCT FROM OLD.generation + 1 OR
       NEW.acceptance_digest IS NULL OR NEW.changed_at <= OLD.changed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5666',
            MESSAGE = 'async V2 switch transition is invalid';
    END IF;
    SELECT run.run_id INTO v_run_id
    FROM platform.async_v1_adoption_runs AS run
    JOIN platform.async_v1_zero_inventory_receipts AS zero_receipt USING (run_id)
    WHERE run.scope_kind OPERATOR(pg_catalog.=) OLD.scope_kind
      AND run.scope_id OPERATOR(pg_catalog.=) OLD.scope_id
      AND run.domain OPERATOR(pg_catalog.=) OLD.domain
      AND run.source_generation OPERATOR(pg_catalog.=) OLD.generation
      AND run.target_generation OPERATOR(pg_catalog.=) NEW.generation
      AND run.state OPERATOR(pg_catalog.=) 'accepted'
      AND run.accepted_at <= NEW.changed_at
      AND run.acceptance_sha256 IS NOT NULL
      AND NEW.acceptance_digest OPERATOR(pg_catalog.=) run.acceptance_sha256
      AND zero_receipt.inventory_sha256 OPERATOR(pg_catalog.=) run.inventory_sha256
      AND zero_receipt.queue_count OPERATOR(pg_catalog.=) 0
      AND zero_receipt.timer_count OPERATOR(pg_catalog.=) 0
      AND zero_receipt.future_count OPERATOR(pg_catalog.=) 0
      AND zero_receipt.running_count OPERATOR(pg_catalog.=) 0
    FOR SHARE OF run, zero_receipt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5667',
            MESSAGE = 'V2 switch requires accepted zero V1 async inventory';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_cutover_guard_v2_async_adoption_v1()
FROM PUBLIC;

CREATE TRIGGER domain_cutovers_async_v1_adoption_guard
BEFORE UPDATE ON platform.domain_cutovers
FOR EACH ROW EXECUTE FUNCTION platform.platform_cutover_guard_v2_async_adoption_v1();
ALTER TABLE platform.domain_cutovers
    ENABLE ALWAYS TRIGGER domain_cutovers_async_v1_adoption_guard;
