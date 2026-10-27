-- +goose Up
CREATE TABLE platform.billing_metered_resource_reconcile_receipts (
    operation_id uuid PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    owner_fact_id uuid NOT NULL,
    idempotency_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    evidence_sha256 bytea NOT NULL,
    evidence_exists boolean NOT NULL,
    evidence_active boolean NOT NULL,
    included_in_opening boolean NOT NULL,
    changed boolean NOT NULL,
    gauge_delta bigint NOT NULL,
    current_amount bigint NOT NULL,
    limit_amount bigint NOT NULL,
    unlimited boolean NOT NULL,
    usage_decision_id uuid,
    usage_entry_id uuid,
    measured_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT billing_metered_resource_reconcile_idempotency_key
        UNIQUE (tenant_id, actor_id, resource_kind, idempotency_key),
    CONSTRAINT billing_metered_resource_reconcile_authority_fk
        FOREIGN KEY (authority_id, tenant_id)
        REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_reconcile_decision_fk
        FOREIGN KEY (usage_decision_id)
        REFERENCES platform.billing_entitlement_decisions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_reconcile_entry_fk
        FOREIGN KEY (usage_entry_id)
        REFERENCES platform.billing_usage_entries (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_reconcile_shape_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND pg_catalog.uuid_extract_version(owner_fact_id) = 7
        AND resource_kind IN ('api_endpoint','analytics_dashboard')
        AND idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
        AND pg_catalog.octet_length(request_sha256) = 32
        AND pg_catalog.octet_length(evidence_sha256) = 32
        AND (evidence_exists OR NOT evidence_active)
        AND (NOT included_in_opening OR evidence_exists AND evidence_active)
        AND gauge_delta BETWEEN -1 AND 1
        AND (changed OR gauge_delta = 0)
        AND current_amount >= 0 AND limit_amount >= 0
        AND (NOT unlimited OR limit_amount = 0)
        AND (gauge_delta <> 0
             OR (current_amount = 0 AND limit_amount = 0 AND NOT unlimited))
        AND pg_catalog.isfinite(measured_at)
        AND pg_catalog.isfinite(recorded_at)
        AND (
            (gauge_delta = 0
             AND usage_decision_id IS NULL AND usage_entry_id IS NULL)
            OR (gauge_delta <> 0
             AND usage_decision_id IS NOT NULL AND usage_entry_id IS NOT NULL)
        )
    )
);

CREATE TRIGGER billing_metered_resource_reconcile_authority_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_metered_resource_reconcile_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
CREATE TRIGGER billing_metered_resource_reconcile_immutable_guard
BEFORE UPDATE OR DELETE ON platform.billing_metered_resource_reconcile_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();

-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource_kind text,
    p_idempotency_key text,
    p_owner_fact_id uuid
)
RETURNS TABLE (
    receipt_found boolean,
    operation_id uuid,
    request_sha256 bytea,
    evidence_sha256 bytea,
    evidence_exists boolean,
    evidence_active boolean,
    included_in_opening boolean,
    changed boolean,
    gauge_delta bigint,
    current_amount bigint,
    limit_amount bigint,
    unlimited boolean,
    recorded_at timestamptz,
    projection_found boolean,
    counted_state text,
    projection_revision bigint,
    projection_updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_receipt platform.billing_metered_resource_reconcile_receipts%ROWTYPE;
    v_projection platform.billing_metered_resources%ROWTYPE;
    v_receipt_found boolean;
    v_projection_found boolean;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF p_tenant_id IS NULL OR p_actor_id IS NULL
       OR p_resource_kind NOT IN ('api_endpoint','analytics_dashboard')
       OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
       OR pg_catalog.uuid_extract_version(p_owner_fact_id) IS DISTINCT FROM 7
       OR NOT EXISTS (
           SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
           WHERE authority_row.authority_id=p_authority_id
             AND authority_row.tenant_id=p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing metered resource reconciliation input is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || '|reconcile|' || p_actor_id::text || '|'
        || p_resource_kind || '|' || p_idempotency_key, 0
    ));
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || '|resource|' || p_resource_kind || '|'
        || p_owner_fact_id::text, 0
    ));

    SELECT receipt_row.* INTO v_receipt
    FROM platform.billing_metered_resource_reconcile_receipts AS receipt_row
    WHERE receipt_row.tenant_id=p_tenant_id
      AND receipt_row.actor_id=p_actor_id
      AND receipt_row.resource_kind=p_resource_kind
      AND receipt_row.idempotency_key=p_idempotency_key
    FOR UPDATE;
    v_receipt_found := FOUND;

    SELECT projection_row.* INTO v_projection
    FROM platform.billing_metered_resources AS projection_row
    WHERE projection_row.tenant_id=p_tenant_id
      AND projection_row.resource_kind=p_resource_kind
      AND projection_row.owner_fact_id=p_owner_fact_id
    FOR UPDATE;
    v_projection_found := FOUND;

    RETURN QUERY SELECT v_receipt_found,
        CASE WHEN v_receipt_found THEN v_receipt.operation_id END,
        CASE WHEN v_receipt_found THEN v_receipt.request_sha256 END,
        CASE WHEN v_receipt_found THEN v_receipt.evidence_sha256 END,
        CASE WHEN v_receipt_found THEN v_receipt.evidence_exists ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.evidence_active ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.included_in_opening ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.changed ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.gauge_delta ELSE 0 END,
        CASE WHEN v_receipt_found THEN v_receipt.current_amount ELSE 0 END,
        CASE WHEN v_receipt_found THEN v_receipt.limit_amount ELSE 0 END,
        CASE WHEN v_receipt_found THEN v_receipt.unlimited ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.recorded_at END,
        v_projection_found,
        CASE WHEN v_projection_found THEN v_projection.counted_state END,
        CASE WHEN v_projection_found THEN v_projection.revision ELSE 0 END,
        CASE WHEN v_projection_found THEN v_projection.updated_at END;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    uuid,uuid,uuid,text,text,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    uuid,uuid,uuid,text,text,uuid
) TO harpy_runtime;
-- +goose StatementEnd
