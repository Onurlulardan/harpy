-- +goose Up
-- Migration version: 20260717010210 UTC
-- +goose StatementBegin
CREATE FUNCTION platform.billing_execute_retention_cycle_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_successor_job_id uuid,
    p_payload_reference text,
    p_payload_sha256 text
)
RETURNS TABLE (
    successor_job_id uuid,
    successor_run_at timestamptz,
    successor_payload_reference text,
    successor_payload_sha256 text,
    successor_dedupe_key text,
    successor_priority integer,
    raw_bodies_redacted bigint,
    raw_records_purged bigint,
    entitlement_records_purged bigint,
    financial_records_purged bigint,
    legal_hold_blocked bigint,
    checkpoint_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored platform.billing_retention_receipts%ROWTYPE;
    now_at timestamptz;
    next_at timestamptz;
    next_key text;
    raw_body_count bigint;
    raw_record_count bigint;
    entitlement_count bigint;
    invoice_count bigint;
    subscription_count bigint;
    hold_count bigint;
    checkpoint_value bigint;
    raw_remaining boolean;
    entitlement_remaining boolean;
    invoice_remaining boolean;
    subscription_remaining boolean;
    receipt_digest bytea;
BEGIN
    now_at := pg_catalog.clock_timestamp();
    raw_body_count := 0;
    raw_record_count := 0;
    entitlement_count := 0;
    invoice_count := 0;
    subscription_count := 0;
    hold_count := 0;
    raw_remaining := false;
    entitlement_remaining := false;
    invoice_remaining := false;
    subscription_remaining := false;
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF NOT EXISTS (
        SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = p_authority_id
          AND authority_row.tenant_id = p_tenant_id
    ) OR pg_catalog.uuid_extract_version(p_job_id) <> 7
       OR pg_catalog.uuid_extract_version(p_attempt_id) <> 7
       OR pg_catalog.uuid_extract_version(p_successor_job_id) <> 7
       OR pg_catalog.octet_length(p_payload_reference) NOT BETWEEN 1 AND 2048
       OR p_payload_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing retention cycle identity is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'billing:retention:' || p_tenant_id::text, 0
    ));
    SELECT * INTO stored
    FROM platform.billing_retention_receipts AS receipt
    WHERE receipt.job_id = p_job_id
    FOR SHARE;
    IF FOUND THEN
        IF stored.authority_id IS DISTINCT FROM p_authority_id
           OR stored.tenant_id IS DISTINCT FROM p_tenant_id
           OR stored.payload_reference IS DISTINCT FROM p_payload_reference
           OR stored.payload_sha256 IS DISTINCT FROM
              pg_catalog.decode(p_payload_sha256, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = 'P5730',
                MESSAGE = 'billing retention job identity was reused';
        END IF;
        RETURN QUERY SELECT stored.successor_job_id, stored.successor_run_at,
            stored.payload_reference, pg_catalog.encode(stored.payload_sha256, 'hex'),
            stored.successor_dedupe_key, 0,
            stored.raw_bodies_redacted, stored.raw_records_purged,
            stored.entitlement_purged, stored.financial_purged,
            stored.legal_hold_blocked, stored.checkpoint_revision;
        RETURN;
    END IF;

    PERFORM 1 FROM platform.billing_legal_holds AS legal_hold
    WHERE legal_hold.tenant_id = p_tenant_id AND legal_hold.state = 'active'
    FOR SHARE;
    IF FOUND THEN
        hold_count := 1;
    ELSE
        INSERT INTO platform.billing_retention_contexts (
            backend_pid, transaction_id, authority_id, tenant_id, job_id, opened_at
        ) VALUES (
            pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(),
            p_authority_id, p_tenant_id, p_job_id, now_at
        );
        SELECT result.raw_bodies_redacted, result.raw_records_purged,
               result.has_remaining
          INTO raw_body_count, raw_record_count, raw_remaining
        FROM platform.billing_purge_raw_retention_v1(
            p_authority_id, p_tenant_id, now_at, 500
        ) AS result;
        SELECT result.records_purged, result.has_remaining
          INTO entitlement_count, entitlement_remaining
        FROM platform.billing_purge_entitlement_retention_v1(
            p_authority_id, p_tenant_id, now_at, 500
        ) AS result;
        SELECT result.records_purged, result.has_remaining
          INTO invoice_count, invoice_remaining
        FROM platform.billing_purge_invoice_retention_v1(
            p_authority_id, p_tenant_id, now_at, 100
        ) AS result;
        SELECT result.records_purged, result.has_remaining
          INTO subscription_count, subscription_remaining
        FROM platform.billing_purge_subscription_retention_v1(
            p_authority_id, p_tenant_id, now_at, 100
        ) AS result;
        DELETE FROM platform.billing_retention_contexts AS retention_context
        WHERE retention_context.backend_pid = pg_catalog.pg_backend_pid()
          AND retention_context.transaction_id = pg_catalog.pg_current_xact_id();
    END IF;

    INSERT INTO platform.billing_retention_checkpoints (
        authority_id, tenant_id, revision, cycle_count, last_job_id,
        last_attempt_id, raw_cutoff, entitlement_cutoff, financial_cutoff,
        last_completed_at
    ) VALUES (
        p_authority_id, p_tenant_id, 1, 1, p_job_id, p_attempt_id,
        now_at - INTERVAL '400 days', now_at - INTERVAL '7 years',
        now_at - INTERVAL '10 years', now_at
    ) ON CONFLICT (tenant_id) DO UPDATE SET
        revision = billing_retention_checkpoints.revision + 1,
        cycle_count = billing_retention_checkpoints.cycle_count + 1,
        last_job_id = EXCLUDED.last_job_id,
        last_attempt_id = EXCLUDED.last_attempt_id,
        raw_cutoff = EXCLUDED.raw_cutoff,
        entitlement_cutoff = EXCLUDED.entitlement_cutoff,
        financial_cutoff = EXCLUDED.financial_cutoff,
        last_completed_at = EXCLUDED.last_completed_at
    RETURNING billing_retention_checkpoints.revision INTO checkpoint_value;

    IF hold_count = 0 AND (
        raw_remaining OR entitlement_remaining
        OR invoice_remaining OR subscription_remaining
    ) THEN
        next_at := now_at + INTERVAL '1 minute';
    ELSE
        next_at := (
            pg_catalog.date_trunc('day', now_at AT TIME ZONE 'UTC')
            + INTERVAL '1 day 3 hours'
        ) AT TIME ZONE 'UTC';
    END IF;
    next_key := 'billing:retention:' || p_successor_job_id::text;
    receipt_digest := pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.jsonb_build_array(
            p_job_id, p_tenant_id, p_successor_job_id, next_at,
            raw_body_count, raw_record_count, entitlement_count,
            invoice_count + subscription_count, hold_count, checkpoint_value
        )::text, 'UTF8'
    ));
    INSERT INTO platform.billing_retention_receipts (
        job_id, authority_id, tenant_id, attempt_id, payload_reference,
        payload_sha256, successor_job_id, successor_run_at,
        successor_dedupe_key, raw_bodies_redacted, raw_records_purged,
        entitlement_purged, financial_purged, legal_hold_blocked,
        checkpoint_revision, receipt_sha256, completed_at
    ) VALUES (
        p_job_id, p_authority_id, p_tenant_id, p_attempt_id,
        p_payload_reference, pg_catalog.decode(p_payload_sha256, 'hex'),
        p_successor_job_id, next_at, next_key, raw_body_count, raw_record_count,
        entitlement_count, invoice_count + subscription_count, hold_count,
        checkpoint_value, receipt_digest, now_at
    );
    RETURN QUERY SELECT p_successor_job_id, next_at, p_payload_reference,
        p_payload_sha256, next_key, 0, raw_body_count, raw_record_count,
        entitlement_count, invoice_count + subscription_count,
        hold_count, checkpoint_value;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_execute_retention_cycle_v1(
    uuid,uuid,uuid,uuid,uuid,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_execute_retention_cycle_v1(
    uuid,uuid,uuid,uuid,uuid,text,text
) TO forgecrud_runtime;
-- +goose StatementEnd
