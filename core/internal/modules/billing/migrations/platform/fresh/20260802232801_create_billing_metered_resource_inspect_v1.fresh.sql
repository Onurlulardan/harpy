-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_inspect_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource_kind text,
    p_mutation_kind text,
    p_idempotency_key text,
    p_owner_fact_id uuid
)
RETURNS TABLE (
    receipt_found boolean,
    operation_id uuid,
    owner_fact_id uuid,
    request_sha256 bytea,
    feature_enabled boolean,
    allowed boolean,
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
    v_receipt platform.billing_metered_resource_receipts%ROWTYPE;
    v_projection platform.billing_metered_resources%ROWTYPE;
    v_receipt_found boolean;
    v_projection_found boolean;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF p_tenant_id IS NULL OR p_actor_id IS NULL
       OR p_resource_kind NOT IN ('api_endpoint','analytics_dashboard')
       OR p_mutation_kind NOT IN (
           'create','update','publish','revalidate','delete','restore'
       )
       OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
       OR pg_catalog.uuid_extract_version(p_owner_fact_id) IS DISTINCT FROM 7
       OR NOT EXISTS (
           SELECT 1
           FROM platform.billing_tenant_authorities AS authority_row
           WHERE authority_row.authority_id = p_authority_id
             AND authority_row.tenant_id = p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing metered resource inspection input is invalid';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || '|receipt|' || p_actor_id::text || '|'
        || p_resource_kind || '|' || p_mutation_kind || '|' || p_idempotency_key,
        0
    ));
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || '|resource|' || p_resource_kind || '|'
        || p_owner_fact_id::text,
        0
    ));

    SELECT receipt_row.*
    INTO v_receipt
    FROM platform.billing_metered_resource_receipts AS receipt_row
    WHERE receipt_row.tenant_id = p_tenant_id
      AND receipt_row.actor_id = p_actor_id
      AND receipt_row.resource_kind = p_resource_kind
      AND receipt_row.mutation_kind = p_mutation_kind
      AND receipt_row.idempotency_key = p_idempotency_key
    FOR UPDATE;
    v_receipt_found := FOUND;

    SELECT projection_row.*
    INTO v_projection
    FROM platform.billing_metered_resources AS projection_row
    WHERE projection_row.tenant_id = p_tenant_id
      AND projection_row.resource_kind = p_resource_kind
      AND projection_row.owner_fact_id = p_owner_fact_id
    FOR UPDATE;
    v_projection_found := FOUND;

    RETURN QUERY SELECT
        v_receipt_found,
        CASE WHEN v_receipt_found THEN v_receipt.operation_id END,
        CASE WHEN v_receipt_found THEN v_receipt.owner_fact_id END,
        CASE WHEN v_receipt_found THEN v_receipt.request_sha256 END,
        CASE WHEN v_receipt_found THEN v_receipt.feature_enabled ELSE false END,
        CASE WHEN v_receipt_found THEN v_receipt.allowed ELSE false END,
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

REVOKE ALL ON FUNCTION platform.billing_metered_resource_inspect_v1(
    uuid,uuid,uuid,text,text,text,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_inspect_v1(
    uuid,uuid,uuid,text,text,text,uuid
) TO harpy_runtime;
-- +goose StatementEnd
