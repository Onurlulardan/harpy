-- +goose Up
-- +goose StatementBegin
CREATE TABLE platform.usage_legal_holds (
    hold_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(hold_id) = 7),
    environment_id uuid NOT NULL,
    tenant_id uuid,
    reason_code text NOT NULL CHECK (reason_code ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    starts_at timestamp(6) with time zone NOT NULL,
    ends_at timestamp(6) with time zone,
    released_at timestamp(6) with time zone,
    created_at timestamp(6) with time zone NOT NULL,
    CHECK (ends_at IS NULL OR ends_at >= starts_at),
    CHECK (released_at IS NULL OR released_at >= starts_at)
);

CREATE TABLE platform.usage_retention_receipts (
    receipt_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    environment_id uuid NOT NULL,
    delete_before timestamp(6) with time zone NOT NULL,
    batch_limit integer NOT NULL CHECK (batch_limit BETWEEN 1 AND 10000),
    deleted_count integer NOT NULL CHECK (deleted_count BETWEEN 0 AND batch_limit),
    observed_at timestamp(6) with time zone NOT NULL
);

-- Retention relations remain owner-only; the adopter is granted only the
-- SECURITY DEFINER routines declared by the final grant migration.

CREATE FUNCTION platform.usage_legal_hold_place_v1(
    p_hold_id uuid, p_environment_id uuid, p_tenant_id uuid, p_reason_code text,
    p_starts_at timestamptz, p_ends_at timestamptz, p_observed_at timestamptz
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_existing platform.usage_legal_holds%ROWTYPE;
BEGIN
    IF p_hold_id IS NULL OR pg_catalog.uuid_extract_version(p_hold_id) <> 7
       OR p_environment_id IS NULL OR p_reason_code !~ '^[a-z][a-z0-9_.-]{0,63}$'
       OR p_starts_at IS NULL OR p_observed_at IS NULL
       OR (p_ends_at IS NOT NULL AND p_ends_at < p_starts_at) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage legal hold is invalid';
    END IF;
    INSERT INTO platform.usage_legal_holds AS target (
        hold_id, environment_id, tenant_id, reason_code, starts_at, ends_at, created_at
    ) VALUES (p_hold_id, p_environment_id, p_tenant_id, p_reason_code,
        p_starts_at, p_ends_at, p_observed_at)
    ON CONFLICT (hold_id) DO NOTHING;
    SELECT * INTO v_existing FROM platform.usage_legal_holds AS holds
    WHERE holds.hold_id = p_hold_id;
    IF v_existing.environment_id IS DISTINCT FROM p_environment_id
       OR v_existing.tenant_id IS DISTINCT FROM p_tenant_id
       OR v_existing.reason_code IS DISTINCT FROM p_reason_code
       OR v_existing.starts_at IS DISTINCT FROM p_starts_at
       OR v_existing.ends_at IS DISTINCT FROM p_ends_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage legal hold conflicts';
    END IF;
    RETURN v_existing.created_at = p_observed_at;
END;
$$;

CREATE FUNCTION platform.usage_legal_hold_release_v1(
    p_hold_id uuid, p_observed_at timestamptz
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
BEGIN
    IF p_hold_id IS NULL OR p_observed_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage legal hold release is invalid';
    END IF;
    UPDATE platform.usage_legal_holds AS holds SET released_at = p_observed_at
    WHERE holds.hold_id = p_hold_id AND holds.released_at IS NULL;
    IF FOUND THEN RETURN true; END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.usage_legal_holds AS holds
        WHERE holds.hold_id = p_hold_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage legal hold not found';
    END IF;
    RETURN false;
END;
$$;

CREATE FUNCTION platform.usage_retention_prune_v1(
    p_receipt_id uuid, p_environment_id uuid, p_delete_before timestamptz,
    p_batch_limit integer, p_observed_at timestamptz
) RETURNS integer LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_existing platform.usage_retention_receipts%ROWTYPE; v_deleted integer;
BEGIN
    IF p_receipt_id IS NULL OR pg_catalog.uuid_extract_version(p_receipt_id) <> 7
       OR p_environment_id IS NULL OR p_observed_at IS NULL
       OR p_batch_limit NOT BETWEEN 1 AND 10000 OR p_delete_before IS NULL
       OR p_delete_before > p_observed_at - interval '31 days' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage retention request is invalid';
    END IF;
    SELECT * INTO v_existing FROM platform.usage_retention_receipts AS receipts
    WHERE receipts.receipt_id = p_receipt_id;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.delete_before IS DISTINCT FROM p_delete_before
           OR v_existing.batch_limit IS DISTINCT FROM p_batch_limit THEN
            RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage retention receipt conflicts';
        END IF;
        RETURN v_existing.deleted_count;
    END IF;
    PERFORM pg_catalog.set_config('forgecrud.usage_retention', 'usage_retention_v1', true);
    WITH candidates AS (
        SELECT facts.fact_id FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id AND facts.occurred_at < p_delete_before
          AND NOT EXISTS (SELECT 1 FROM platform.usage_legal_holds AS holds
              WHERE holds.environment_id = facts.environment_id AND holds.released_at IS NULL
                AND (holds.tenant_id IS NULL OR holds.tenant_id = facts.tenant_id)
                AND holds.starts_at <= p_observed_at
                AND (holds.ends_at IS NULL OR holds.ends_at >= p_observed_at))
        ORDER BY facts.occurred_at, facts.fact_id LIMIT p_batch_limit FOR UPDATE SKIP LOCKED
    ) DELETE FROM platform.usage_request_facts AS facts
      USING candidates WHERE facts.fact_id = candidates.fact_id;
    GET DIAGNOSTICS v_deleted = ROW_COUNT;
    INSERT INTO platform.usage_retention_receipts (
        receipt_id, environment_id, delete_before, batch_limit, deleted_count, observed_at
    ) VALUES (p_receipt_id, p_environment_id, p_delete_before,
        p_batch_limit, v_deleted, p_observed_at);
    RETURN v_deleted;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_legal_hold_place_v1(
uuid,uuid,uuid,text,timestamptz,timestamptz,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_legal_hold_release_v1(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_retention_prune_v1(
uuid,uuid,timestamptz,integer,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
