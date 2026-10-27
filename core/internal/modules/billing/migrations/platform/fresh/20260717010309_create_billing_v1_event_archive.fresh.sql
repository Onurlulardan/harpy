-- +goose Up
-- Migration version: 20260717010309 UTC
CREATE TABLE platform.billing_v1_event_archive (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    source_event_id uuid NOT NULL,
    subscription_id uuid,
    source_subscription_id uuid,
    event_type text NOT NULL,
    amount_cents bigint NOT NULL,
    currency text NOT NULL,
    description text NOT NULL,
    metadata jsonb,
    created_at timestamptz NOT NULL,
    source_sha256 bytea NOT NULL,
    adoption_operation_id uuid NOT NULL,
    adoption_run_id uuid NOT NULL,
    archived_at timestamptz NOT NULL,
    CONSTRAINT billing_v1_event_archive_pkey PRIMARY KEY (id),
    CONSTRAINT billing_v1_event_archive_source_key
        UNIQUE (tenant_id, source_event_id),
    CONSTRAINT billing_v1_event_archive_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_event_archive_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_event_archive_run_fk FOREIGN KEY (
        adoption_operation_id, adoption_run_id
    ) REFERENCES platform.billing_v1_adoption_runs (operation_id, run_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_v1_event_archive_shape_check CHECK (
        event_type IN (
            'subscription_created', 'subscription_upgraded',
            'subscription_downgraded', 'subscription_cancelled',
            'subscription_resumed', 'subscription_renewed',
            'subscription_expired', 'trial_started', 'trial_ended',
            'plan_changed', 'coupon_applied', 'coupon_removed',
            'payment_succeeded', 'payment_failed'
        )
        AND pg_catalog.char_length(currency) BETWEEN 1 AND 3
        AND (metadata IS NULL OR
            pg_catalog.octet_length(metadata::text) <= 65536)
        AND pg_catalog.octet_length(source_sha256) = 32
    )
);
CREATE TRIGGER billing_v1_event_archive_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_event_archive
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_v1_event_archive ENABLE ALWAYS TRIGGER
    billing_v1_event_archive_authority_guard;
CREATE TRIGGER billing_v1_event_archive_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_event_archive
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_event_archive ENABLE ALWAYS TRIGGER
    billing_v1_event_archive_immutable_guard;

-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_event_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_id uuid;
    v_sha bytea;
BEGIN
    IF p_operation_id IS NULL OR p_run_id IS NULL OR p_record IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 event input is invalid';
    END IF;
    v_id := (p_record->>'id')::uuid;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'tenant' AND run.state = 'planned' AND run.eligible;
    IF (p_record->>'tenant_id')::uuid IS DISTINCT FROM v_run.scope_id THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 event tenant is invalid';
    END IF;
    INSERT INTO platform.billing_v1_event_archive (
        id, authority_id, tenant_id, source_event_id, subscription_id,
        source_subscription_id, event_type, amount_cents, currency,
        description, metadata, created_at, source_sha256,
        adoption_operation_id, adoption_run_id, archived_at
    ) VALUES (
        v_id, v_run.authority_id, v_run.scope_id,
        (p_record->>'source_id')::uuid,
        NULLIF(p_record->>'subscription_id', '')::uuid,
        NULLIF(p_record->>'source_subscription_id', '')::uuid,
        p_record->>'event_type', (p_record->>'amount_cents')::bigint,
        p_record->>'currency', p_record->>'description',
        CASE WHEN p_record ? 'metadata' THEN p_record->'metadata' ELSE NULL END,
        (p_record->>'created_at')::timestamptz,
        pg_catalog.decode(p_record->>'source_sha256', 'hex'),
        p_operation_id, p_run_id, v_run.snapshot_at
    );
    v_sha := platform.billing_v1_adoption_tenant_fingerprint_v1(
        'legacy_event', v_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'billing_events', p_record->>'source_id',
        'legacy_event', v_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_event_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
