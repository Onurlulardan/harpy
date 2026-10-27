-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_provider_command_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF OLD.actor_id IS NULL THEN
        RETURN NEW;
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.provider_account_id IS DISTINCT FROM OLD.provider_account_id
       OR NEW.provider_object_mapping_id IS DISTINCT FROM
          OLD.provider_object_mapping_id
       OR NEW.subscription_id IS DISTINCT FROM OLD.subscription_id
       OR NEW.command_kind IS DISTINCT FROM OLD.command_kind
       OR NEW.effect_key IS DISTINCT FROM OLD.effect_key
       OR NEW.request_sha256 IS DISTINCT FROM OLD.request_sha256
       OR NEW.requested_at IS DISTINCT FROM OLD.requested_at
       OR NEW.actor_id IS DISTINCT FROM OLD.actor_id
       OR NEW.revision <> OLD.revision + 1
       OR (OLD.state IN ('succeeded','failed'))
       OR NOT (
           (OLD.state IN ('queued','running') AND NEW.state = OLD.state
            AND NEW.active_job_id IS NOT DISTINCT FROM OLD.active_job_id
            AND NEW.finalizer_job_id IS DISTINCT FROM OLD.finalizer_job_id)
           OR (OLD.state IN ('queued','running')
               AND NEW.state = 'reconcile_required'
               AND NEW.active_job_id IS DISTINCT FROM OLD.active_job_id
               AND NEW.finalizer_job_id IS DISTINCT FROM OLD.finalizer_job_id)
           OR (OLD.state IN ('queued','running','reconcile_required')
               AND NEW.state IN ('succeeded','failed')
               AND NEW.active_job_id IS NOT DISTINCT FROM OLD.active_job_id
               AND NEW.finalizer_job_id IS NOT DISTINCT FROM OLD.finalizer_job_id)
           OR (OLD.state = 'reconcile_required'
               AND NEW.state = 'reconcile_required'
               AND NEW.active_job_id IS NOT DISTINCT FROM OLD.active_job_id)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing provider command transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_guard_provider_command_v1() FROM PUBLIC;

CREATE TRIGGER billing_provider_commands_transition_guard
BEFORE UPDATE ON platform.billing_provider_commands
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_provider_command_v1();
ALTER TABLE platform.billing_provider_commands
    ENABLE ALWAYS TRIGGER billing_provider_commands_transition_guard;
-- +goose StatementEnd
