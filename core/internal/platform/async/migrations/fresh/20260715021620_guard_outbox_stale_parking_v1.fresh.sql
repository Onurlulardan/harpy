-- +goose Up
CREATE OR REPLACE TRIGGER outbox_transition_guard
BEFORE UPDATE ON platform.outbox
FOR EACH ROW WHEN (NEW.parked_at IS NULL)
EXECUTE FUNCTION platform.platform_async_guard_outbox_transition_v1();
ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard;

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_outbox_park_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'outbox' OR
       SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' OR
       CURRENT_USER OPERATOR(pg_catalog.<>) 'forgecrud_owner' OR
       OLD.parked_at IS NOT NULL OR NEW.parked_at IS NULL OR
       NEW.parked_code IS DISTINCT FROM 'OUTBOX_AUTHORITY_STALE' OR
       NEW.parked_current_group_generation <
           OLD.authority_group_generation OR
       NEW.parked_current_cutover_generation < OLD.cutover_generation OR
       (NEW.parked_current_group_generation =
            OLD.authority_group_generation AND
        NEW.parked_current_cutover_generation = OLD.cutover_generation) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5650',
            MESSAGE = 'outbox stale park invocation is invalid';
    END IF;

    PERFORM platform.platform_async_assert_current_authority_v1(
        OLD.scope_kind, OLD.scope_id, OLD.authority_group_id,
        NEW.parked_current_group_generation, OLD.authority_domain,
        NEW.parked_current_cutover_generation
    );

    IF NEW.delivery_state IS DISTINCT FROM 'pending' OR
       NEW.lease_owner_id IS NOT NULL OR
       NEW.lease_heartbeat_at IS NOT NULL OR NEW.lease_until IS NOT NULL OR
       NOT (
           (OLD.delivery_state = 'pending' AND
            OLD.lease_owner_id IS NULL AND OLD.lease_heartbeat_at IS NULL AND
            OLD.lease_until IS NULL) OR
           (OLD.delivery_state = 'publishing' AND
            OLD.lease_until <= v_now)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5651',
            MESSAGE = 'outbox stale park lease is not eligible';
    END IF;

    IF ROW(
        NEW.event_id, NEW.tenant_id, NEW.scope_kind, NEW.scope_id,
        NEW.authority_group_id, NEW.authority_group_generation,
        NEW.authority_domain, NEW.cutover_generation, NEW.writer_origin,
        NEW.envelope_contract_version, NEW.event_type, NEW.event_version,
        NEW.aggregate_type, NEW.aggregate_id, NEW.aggregate_version,
        NEW.event_sequence, NEW.occurred_at, NEW.actor_id,
        NEW.correlation_id, NEW.causation_id, NEW.idempotency_key,
        NEW.business_effect_idempotency_key, NEW.payload_canonical,
        NEW.payload_sha256, NEW.payload_schema_sha256, NEW.payload_redacted_at,
        NEW.delivery_attempts, NEW.lease_generation, NEW.next_attempt_at,
        NEW.confirmed_at, NEW.delivered_at, NEW.confirm_sha256,
        NEW.last_error_code, NEW.last_error_sha256, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.event_id, OLD.tenant_id, OLD.scope_kind, OLD.scope_id,
        OLD.authority_group_id, OLD.authority_group_generation,
        OLD.authority_domain, OLD.cutover_generation, OLD.writer_origin,
        OLD.envelope_contract_version, OLD.event_type, OLD.event_version,
        OLD.aggregate_type, OLD.aggregate_id, OLD.aggregate_version,
        OLD.event_sequence, OLD.occurred_at, OLD.actor_id,
        OLD.correlation_id, OLD.causation_id, OLD.idempotency_key,
        OLD.business_effect_idempotency_key, OLD.payload_canonical,
        OLD.payload_sha256, OLD.payload_schema_sha256, OLD.payload_redacted_at,
        OLD.delivery_attempts, OLD.lease_generation, OLD.next_attempt_at,
        OLD.confirmed_at, OLD.delivered_at, OLD.confirm_sha256,
        OLD.last_error_code, OLD.last_error_sha256, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5652',
            MESSAGE = 'outbox stale park changed immutable state';
    END IF;

    NEW.parked_at := v_now;
    NEW.updated_at := v_now;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_outbox_park_v1()
FROM PUBLIC;

CREATE TRIGGER outbox_park_guard
BEFORE UPDATE ON platform.outbox
FOR EACH ROW WHEN (NEW.parked_at IS NOT NULL)
EXECUTE FUNCTION platform.platform_async_guard_outbox_park_v1();
ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_park_guard;
