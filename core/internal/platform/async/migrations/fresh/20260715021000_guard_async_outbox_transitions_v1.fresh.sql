-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_outbox_transition_v1()
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
       SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5620',
            MESSAGE = 'async outbox mutation authority is invalid';
    END IF;

    IF ROW(
        NEW.event_id, NEW.tenant_id, NEW.scope_kind, NEW.scope_id,
        NEW.authority_group_id, NEW.authority_group_generation,
        NEW.authority_domain, NEW.cutover_generation, NEW.writer_origin,
        NEW.envelope_contract_version, NEW.event_type, NEW.event_version,
        NEW.aggregate_type, NEW.aggregate_id, NEW.aggregate_version,
        NEW.event_sequence, NEW.occurred_at, NEW.actor_id,
        NEW.correlation_id, NEW.causation_id, NEW.idempotency_key,
        NEW.business_effect_idempotency_key, NEW.payload_sha256,
        NEW.payload_schema_sha256, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.event_id, OLD.tenant_id, OLD.scope_kind, OLD.scope_id,
        OLD.authority_group_id, OLD.authority_group_generation,
        OLD.authority_domain, OLD.cutover_generation, OLD.writer_origin,
        OLD.envelope_contract_version, OLD.event_type, OLD.event_version,
        OLD.aggregate_type, OLD.aggregate_id, OLD.aggregate_version,
        OLD.event_sequence, OLD.occurred_at, OLD.actor_id,
        OLD.correlation_id, OLD.causation_id, OLD.idempotency_key,
        OLD.business_effect_idempotency_key, OLD.payload_sha256,
        OLD.payload_schema_sha256, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5621',
            MESSAGE = 'async outbox envelope is immutable';
    END IF;

    NEW.updated_at := v_now;

    IF CURRENT_USER OPERATOR(pg_catalog.=) 'forgecrud_owner' AND
       OLD.delivery_state OPERATOR(pg_catalog.=) 'delivered' AND
       NEW.delivery_state OPERATOR(pg_catalog.=) 'delivered' AND
       OLD.payload_canonical IS NOT NULL AND NEW.payload_canonical IS NULL AND
       OLD.payload_redacted_at IS NULL AND
       OLD.delivered_at <=
           (v_now OPERATOR(pg_catalog.-) INTERVAL '30 days') AND
       ROW(NEW.delivery_attempts, NEW.lease_owner_id, NEW.lease_generation,
           NEW.lease_heartbeat_at, NEW.lease_until, NEW.next_attempt_at,
           NEW.confirmed_at, NEW.delivered_at, NEW.confirm_sha256,
           NEW.last_error_code, NEW.last_error_sha256)
       IS NOT DISTINCT FROM
       ROW(OLD.delivery_attempts, OLD.lease_owner_id, OLD.lease_generation,
           OLD.lease_heartbeat_at, OLD.lease_until, OLD.next_attempt_at,
           OLD.confirmed_at, OLD.delivered_at, OLD.confirm_sha256,
           OLD.last_error_code, OLD.last_error_sha256) THEN
        NEW.payload_redacted_at := v_now;
        RETURN NEW;
    END IF;

    PERFORM platform.platform_async_assert_current_authority_v1(
        OLD.scope_kind, OLD.scope_id,
        OLD.authority_group_id, OLD.authority_group_generation,
        OLD.authority_domain, OLD.cutover_generation
    );

    IF OLD.delivery_state OPERATOR(pg_catalog.=) 'pending' AND
       NEW.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
       NEW.delivery_attempts OPERATOR(pg_catalog.=)
           (OLD.delivery_attempts OPERATOR(pg_catalog.+) 1) AND
       NEW.lease_generation OPERATOR(pg_catalog.=)
           (OLD.lease_generation OPERATOR(pg_catalog.+) 1) AND
       OLD.next_attempt_at <= v_now AND NEW.lease_owner_id IS NOT NULL AND
       ROW(NEW.payload_canonical, NEW.payload_redacted_at,
           NEW.next_attempt_at, NEW.confirmed_at, NEW.delivered_at,
           NEW.confirm_sha256, NEW.last_error_code, NEW.last_error_sha256)
       IS NOT DISTINCT FROM
       ROW(OLD.payload_canonical, OLD.payload_redacted_at,
           OLD.next_attempt_at, OLD.confirmed_at, OLD.delivered_at,
           OLD.confirm_sha256, OLD.last_error_code, OLD.last_error_sha256) THEN
        IF (OLD.event_sequence OPERATOR(pg_catalog.>) 1 AND NOT EXISTS (
            SELECT 1 FROM platform.outbox AS predecessor
            WHERE predecessor.scope_kind OPERATOR(pg_catalog.=) OLD.scope_kind
              AND predecessor.scope_id OPERATOR(pg_catalog.=) OLD.scope_id
              AND predecessor.aggregate_type OPERATOR(pg_catalog.=)
                  OLD.aggregate_type
              AND predecessor.aggregate_id OPERATOR(pg_catalog.=) OLD.aggregate_id
              AND predecessor.event_sequence OPERATOR(pg_catalog.=)
                  (OLD.event_sequence OPERATOR(pg_catalog.-) 1)
              AND predecessor.delivery_state OPERATOR(pg_catalog.=) 'delivered'
        )) OR EXISTS (
            SELECT 1 FROM platform.outbox AS prior
            WHERE prior.scope_kind OPERATOR(pg_catalog.=) OLD.scope_kind
              AND prior.scope_id OPERATOR(pg_catalog.=) OLD.scope_id
              AND prior.aggregate_type OPERATOR(pg_catalog.=) OLD.aggregate_type
              AND prior.aggregate_id OPERATOR(pg_catalog.=) OLD.aggregate_id
              AND prior.event_sequence OPERATOR(pg_catalog.<) OLD.event_sequence
              AND prior.delivery_state OPERATOR(pg_catalog.<>) 'delivered'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5623',
                MESSAGE = 'async outbox stream predecessor is not delivered';
        END IF;
        NEW.lease_heartbeat_at := v_now;
        NEW.lease_until := v_now OPERATOR(pg_catalog.+) INTERVAL '60 seconds';
        RETURN NEW;
    END IF;

    IF OLD.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
       NEW.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
       OLD.lease_until > v_now AND
       NEW.lease_until > OLD.lease_until AND
       ROW(NEW.delivery_attempts, NEW.lease_owner_id, NEW.lease_generation,
           NEW.payload_canonical, NEW.payload_redacted_at,
           NEW.next_attempt_at, NEW.confirmed_at, NEW.delivered_at,
           NEW.confirm_sha256, NEW.last_error_code, NEW.last_error_sha256)
       IS NOT DISTINCT FROM
       ROW(OLD.delivery_attempts, OLD.lease_owner_id, OLD.lease_generation,
           OLD.payload_canonical, OLD.payload_redacted_at,
           OLD.next_attempt_at, OLD.confirmed_at, OLD.delivered_at,
           OLD.confirm_sha256, OLD.last_error_code, OLD.last_error_sha256) THEN
        NEW.lease_heartbeat_at := v_now;
        NEW.lease_until := v_now OPERATOR(pg_catalog.+) INTERVAL '60 seconds';
        RETURN NEW;
    END IF;

    IF OLD.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
       NEW.delivery_state OPERATOR(pg_catalog.=) 'pending' AND
       NEW.delivery_attempts OPERATOR(pg_catalog.=) OLD.delivery_attempts AND
       NEW.lease_generation OPERATOR(pg_catalog.=) OLD.lease_generation AND
       NEW.lease_owner_id IS NULL AND NEW.lease_heartbeat_at IS NULL AND
       NEW.lease_until IS NULL AND NEW.next_attempt_at IS NOT NULL AND
       NEW.last_error_code IS NOT NULL AND NEW.last_error_sha256 IS NOT NULL AND
       ROW(NEW.payload_canonical, NEW.payload_redacted_at,
           NEW.confirmed_at, NEW.delivered_at, NEW.confirm_sha256)
       IS NOT DISTINCT FROM
       ROW(OLD.payload_canonical, OLD.payload_redacted_at,
           OLD.confirmed_at, OLD.delivered_at, OLD.confirm_sha256) THEN
        IF NEW.next_attempt_at < v_now THEN
            NEW.next_attempt_at := v_now;
        END IF;
        RETURN NEW;
    END IF;

    IF OLD.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
       NEW.delivery_state OPERATOR(pg_catalog.=) 'delivered' AND
       OLD.lease_until > v_now AND
       NEW.delivery_attempts OPERATOR(pg_catalog.=) OLD.delivery_attempts AND
       NEW.lease_generation OPERATOR(pg_catalog.=) OLD.lease_generation AND
       NEW.lease_owner_id IS NULL AND NEW.lease_heartbeat_at IS NULL AND
       NEW.lease_until IS NULL AND NEW.confirm_sha256 IS NOT NULL AND
       NEW.next_attempt_at IS NOT DISTINCT FROM OLD.next_attempt_at AND
       NEW.payload_canonical IS NOT DISTINCT FROM OLD.payload_canonical AND
       NEW.payload_redacted_at IS NOT DISTINCT FROM OLD.payload_redacted_at THEN
        NEW.confirmed_at := v_now;
        NEW.delivered_at := v_now;
        NEW.last_error_code := NULL;
        NEW.last_error_sha256 := NULL;
        RETURN NEW;
    END IF;

    RAISE EXCEPTION USING ERRCODE = 'P5622',
        MESSAGE = 'async outbox transition is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_outbox_transition_v1()
FROM PUBLIC;

CREATE TRIGGER outbox_transition_guard
BEFORE UPDATE ON platform.outbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_outbox_transition_v1();
ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard;
