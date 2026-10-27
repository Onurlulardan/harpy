-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_v1_adoption_run_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_changed jsonb;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'async_v1_adoption_runs' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5630',
            MESSAGE = 'V1 async adoption run mutation is forbidden';
    END IF;

    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;

    IF OLD.state OPERATOR(pg_catalog.=) 'inventory' AND
       NEW.state OPERATOR(pg_catalog.=) 'disposing' AND
       v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY['state']::text[]) THEN
        RETURN NEW;
    END IF;
    IF OLD.state OPERATOR(pg_catalog.=) 'disposing' AND
       NEW.state OPERATOR(pg_catalog.=) 'accepted' AND
       NEW.accepted_at IS NOT NULL AND NEW.acceptance_sha256 IS NOT NULL AND
       v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'state', 'accepted_at', 'acceptance_sha256'
       ]::text[]) THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = 'P5631',
        MESSAGE = 'V1 async adoption run transition is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_v1_adoption_run_v1()
FROM PUBLIC;

CREATE TRIGGER async_v1_adoption_runs_transition_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_adoption_runs
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_v1_adoption_run_v1();
ALTER TABLE platform.async_v1_adoption_runs
    ENABLE ALWAYS TRIGGER async_v1_adoption_runs_transition_guard;

CREATE TRIGGER async_v1_adoption_items_immutable_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_adoption_items
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.async_v1_adoption_items
    ENABLE ALWAYS TRIGGER async_v1_adoption_items_immutable_guard;

CREATE TRIGGER async_v1_adoption_dispositions_immutable_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_adoption_dispositions
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.async_v1_adoption_dispositions
    ENABLE ALWAYS TRIGGER async_v1_adoption_dispositions_immutable_guard;

CREATE TRIGGER async_v1_adoption_job_bindings_immutable_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_adoption_job_bindings
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.async_v1_adoption_job_bindings
    ENABLE ALWAYS TRIGGER async_v1_adoption_job_bindings_immutable_guard;

CREATE TRIGGER async_v1_restart_fences_immutable_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_restart_fence_receipts
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.async_v1_restart_fence_receipts
    ENABLE ALWAYS TRIGGER async_v1_restart_fences_immutable_guard;

CREATE TRIGGER async_v1_zero_inventory_immutable_guard
BEFORE UPDATE OR DELETE ON platform.async_v1_zero_inventory_receipts
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.async_v1_zero_inventory_receipts
    ENABLE ALWAYS TRIGGER async_v1_zero_inventory_immutable_guard;
